package main

import (
	"encoding/json"
	"fmt"
	"light-stemcell-builder/builder"
	"light-stemcell-builder/ec2/ec2ami"
	"light-stemcell-builder/ec2/ec2cli"
	"log"
	"os"
)

type Config struct {
	AccessKey    string        `json:"access_key"`
	SecretKey    string        `json:"secret_key"`
	BucketName   string        `json:"bucket_name"`
	Region       string        `json:"region"`
	StemcellPath string        `json:"stemcell_path"`
	OutputPath   string        `json:"output_path"`
	CopyDests    []string      `json:"copy_dests"`
	AmiConfig    ec2ami.Config `json:"ami_configuration"`
}

func validateConfig(config *Config) error {
	if config.AccessKey == "" {
		return fmt.Errorf("access_key can't be empty")
	}
	if config.SecretKey == "" {
		return fmt.Errorf("secret_key can't be empty")
	}
	if config.BucketName == "" {
		return fmt.Errorf("bucket_name can't be empty")
	}
	if config.Region == "" {
		return fmt.Errorf("region can't be empty")
	}
	if config.StemcellPath == "" {
		return fmt.Errorf("stemcell_path can't be empty")
	}
	if config.OutputPath == "" {
		return fmt.Errorf("output_path can't be empty")
	}

	err := config.AmiConfig.Validate()
	if err != nil {
		return fmt.Errorf("Error validating ami_configuration: %s", err.Error())
	}
	return nil
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	if len(os.Args) != 2 {
		logger.Fatalln("Usage: light-stemcell-builder path_to_config.json")
	}
	pathToConfig := os.Args[1]

	configFile, err := os.Open(pathToConfig)
	defer func() {
		err = configFile.Close()
		if err != nil {
			logger.Fatalf("Error closing config file: %s\n", err.Error())
		}
	}()

	if err != nil {
		logger.Fatalf("Error opening config file: %s\n", err.Error())
	}

	config := &Config{}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(config); err != nil {
		logger.Fatalf("Can't parse %s config file. Error is : %s\n", pathToConfig, err.Error())
	}

	config.AmiConfig.Region = config.Region
	err = validateConfig(config)

	var awsConfig = builder.AwsConfig{
		AccessKey:  config.AccessKey,
		SecretKey:  config.SecretKey,
		BucketName: config.BucketName,
		Region:     config.Region,
	}

	aws := &ec2cli.EC2Cli{}
	stemcellBuilder := builder.New(logger, aws, awsConfig, config.AmiConfig)

	stemcellPath, amis, err := stemcellBuilder.BuildLightStemcell(config.StemcellPath, config.OutputPath, config.CopyDests)
	if err != nil {
		logger.Fatalf("Error during stemcell builder: %s\n", err)
	}
	amiJson, err := json.Marshal(amis)
	if err != nil {
		logger.Printf("Error output encoding: %s\n", err)
	}
	logger.Printf("Created AMIs:\n%s", amiJson)

	logger.Printf("Output saved to: %s\n", stemcellPath)
}