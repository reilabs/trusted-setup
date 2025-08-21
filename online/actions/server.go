package actions

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/offline/phase1"
	"github.com/reilabs/trusted-setup/offline/r1cs"
	server_config "github.com/reilabs/trusted-setup/online/config"
	"github.com/reilabs/trusted-setup/online/contribution"
	"github.com/reilabs/trusted-setup/online/server"
	"github.com/reilabs/trusted-setup/online/server/ceremony_service"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
	"github.com/reilabs/trusted-setup/online/server/coordinator"
	"github.com/reilabs/trusted-setup/online/storage"
	"github.com/reilabs/trusted-setup/utils/randomness"
)

func Server(_ context.Context, cmd *cli.Command) error {
	configFilePath := cmd.String("config")

	log.Printf("Loading config file: %s", configFilePath)
	config, err := server_config.New(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	ccs, err := r1cs.FromFile(config.R1cs)
	if err != nil {
		return err
	}

	p1, err := phase1.FromFile(config.Phase1)
	if err != nil {
		return err
	}

	beaconProvider, err := randomness.New()
	if err != nil {
		return err
	}
	beacon := beaconProvider.GetBeacon()
	log.Printf("Beacon: %x", beacon)

	store := storage.NewTmpfs(config.CeremonyName)

	log.Print("Initializing Phase 2")
	last, err := contribution.New(p1, ccs, store, beacon)
	if err != nil {
		return err
	}

	service := ceremony_service.New(
		config.CeremonyName,
		coordinator.New(
			last,
			contributors_manager.New(),
		),
	)

	s := server.New(service)

	err = s.Start(config.Host, config.Port)
	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Press Ctrl+C to end Ceremony and generate Keys")
	<-sigs
	s.Stop()

	if last.GetCount() > 0 {
		log.Printf("Generating keys out of %d contributions...\n", last.GetCount())
		_, _, err = last.ExtractKeys()
	} else {
		log.Printf("No contributions received")
	}

	log.Println("Artifacts generated in the ceremony:")
	files, err := store.List()
	if err != nil {
		return err
	}
	for _, file := range files {
		log.Println("\t" + file)
	}

	return err
}
