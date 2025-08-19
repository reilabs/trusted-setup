package actions

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/offline/phase1"
	offline_phase2 "github.com/reilabs/trusted-setup/offline/phase2"
	"github.com/reilabs/trusted-setup/offline/r1cs"
	server_config "github.com/reilabs/trusted-setup/online/config"
	"github.com/reilabs/trusted-setup/online/contribution"
	"github.com/reilabs/trusted-setup/online/server"
	"github.com/reilabs/trusted-setup/online/server/ceremony_service"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
	"github.com/reilabs/trusted-setup/online/server/coordinator"
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

	log.Print("Initializing Phase 2")
	last := contribution.New(p1, ccs, beaconProvider.GetBeacon())

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
	fmt.Println("Press Ctrl+C to end Ceremony and generate Keys")
	<-sigs
	s.Stop()

	// TODO: this is temporary, keys will go to S3
	pkTemp, err := getTempFilePath("pk")
	if err != nil {
		return err
	}
	vkTemp, err := getTempFilePath("vk")
	if err != nil {
		return err
	}
	fmt.Printf("Generating keys out of %d contributions...", last.GetCount())
	pk, vk := last.ExtractKeys()
	return offline_phase2.PkVkToFile(pk, pkTemp, vk, vkTemp)
}

func getTempFilePath(pattern string) (string, error) {
	tempFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	// Close immediately because we're not writing to these files, we just need paths
	err = tempFile.Close()
	if err != nil {
		log.Printf("error closing %s", tempFile.Name())
	}

	return tempFile.Name(), nil
}
