package options

import (
	"fmt"
	"os"
	"strings"
)

var (
	CIVO_REGION        = "CIVO_REGION"
	CIVO_INSTANCE_TYPE = "CIVO_INSTANCE_TYPE"
	CIVO_DISK_IMAGE    = "CIVO_DISK_IMAGE"
)

type Options struct {
	DiskImage     string
	DiskSizeGB    int
	MachineFolder string
	MachineID     string
	MachineType   string
	Region        string
}

func ConfigFromEnv() (Options, error) {
	return Options{
		MachineType: os.Getenv(CIVO_INSTANCE_TYPE),
		DiskImage:   os.Getenv(CIVO_DISK_IMAGE),
		Region:      os.Getenv(CIVO_REGION),
	}, nil
}

func FromEnv(init, withFolder bool) (*Options, error) {
	retOptions := &Options{}

	var err error

	retOptions.MachineType, err = fromEnvOrError("CIVO_INSTANCE_TYPE")
	if err != nil {
		return nil, err
	}

	retOptions.DiskImage, err = fromEnvOrError("CIVO_DISK_IMAGE")
	if err != nil {
		return nil, err
	}

	retOptions.Region, err = fromEnvOrError("CIVO_REGION")
	if err != nil {
		return nil, err
	}

	// Return early if we're just doing init.
	if init {
		return retOptions, nil
	}

	if err := loadMachineOptions(retOptions, withFolder); err != nil {
		return nil, err
	}

	return retOptions, nil
}

func loadMachineOptions(opts *Options, withFolder bool) error {
	machineID, err := fromEnvOrError("MACHINE_ID")
	if err != nil {
		return err
	}
	// prefix with devsy-
	if !strings.HasPrefix(machineID, "devsy-") {
		machineID = "devsy-" + machineID
	}
	opts.MachineID = machineID

	if withFolder {
		opts.MachineFolder, err = fromEnvOrError("MACHINE_FOLDER")
		if err != nil {
			return err
		}
	}

	return nil
}

func fromEnvOrError(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf(
			"couldn't find option %s in environment, please make sure %s is defined",
			name,
			name,
		)
	}

	return val, nil
}
