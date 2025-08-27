package actions

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
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

	logFile, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer func(logFile *os.File) {
		err = logFile.Close()
		if err != nil {
			log.Printf("Error closing log file writer: %v", err)
		}
	}(logFile)
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006/01/02 15:04:05"}
	ceremonyLogger := zerolog.New(io.MultiWriter(consoleWriter, logFile)).With().Timestamp().Logger()

	beaconProvider, err := randomness.New()
	if err != nil {
		return err
	}
	beacon := beaconProvider.GetBeacon()
	ceremonyLogger.Info().Hex("beacon", beacon).Send()

	var store storage.Storage
	if !config.UseS3 {
		log.Print("Ceremony artifacts will be stored in tmpfs")
		store = storage.NewTmpfs(config.CeremonyName)
	} else {
		log.Println("Ceremony artifacts will be stored in AWS S3")
		var s3Opts []storage.S3Option
		if config.S3Bucket != "" {
			log.Printf("\tbucket: %s", config.S3Bucket)
			s3Opts = append(s3Opts, storage.WithBucket(config.S3Bucket))
		}
		if config.S3Region != "" {
			log.Printf("\tregion: %s", config.S3Region)
			s3Opts = append(s3Opts, storage.WithRegion(config.S3Region))
		}
		if config.S3Profile != "" {
			log.Printf("\tprofile: %s", config.S3Profile)
			s3Opts = append(s3Opts, storage.WithProfile(config.S3Profile))
		}
		if config.S3CredentialsFile != "" {
			log.Printf("\tcredentials file: %s", config.S3CredentialsFile)
			s3Opts = append(s3Opts, storage.WithCredentialsFile(config.S3CredentialsFile))
		}
		store, err = storage.NewS3(s3Opts...)
		if err != nil {
			return err
		}
	}

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
		&ceremonyLogger,
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

	_, err = logFile.Seek(0, 0)
	if err != nil {
		log.Printf("Rewinding log file failed")
	}
	if _, err = store.Save("log", logFile); err != nil {
		log.Printf("Storing ceremony log failed")
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
