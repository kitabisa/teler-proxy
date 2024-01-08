package common

import "testing"

func TestOptions_Validate_ValidConfig(t *testing.T) {
	opt := &Options{
		Port:        &Port{Server: 8080},
		Destination: "example.com",
		Config: &Config{
			Path:   "config.yaml",
			Format: "yaml",
		},
		TLS: nil,
	}

	err := opt.Validate()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func TestOptions_Validate_EmptyDestination(t *testing.T) {
	opt := &Options{
		Port:        &Port{Server: 8080},
		Destination: "",
		Config: &Config{
			Path:   "config.yaml",
			Format: "yaml",
		},
		TLS: nil,
	}

	err := opt.Validate()
	if err != ErrDestAddressEmpty {
		t.Errorf("Expected ErrDestAddressEmpty, but got: %v", err)
	}
}

func TestOptions_Validate_InvalidConfigFormat(t *testing.T) {
	opt := &Options{
		Port:        &Port{Server: 8080},
		Destination: "example.com",
		Config: &Config{
			Path:   "config.json",
			Format: "xml",
		},
		TLS: nil,
	}

	err := opt.Validate()
	if err != ErrCfgFileFormatInv {
		t.Errorf("Expected ErrCfgFileFormatInv, but got: %v", err)
	}
}

func TestOptions_Validate_MissingConfigPathAndFormat(t *testing.T) {
	opt := &Options{
		Port:        &Port{Server: 8080},
		Destination: "example.com",
		Config:      &Config{},
		TLS:         nil,
	}

	err := opt.Validate()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func TestOptions_Validate_NilOptions(t *testing.T) {
	opt := &Options{}

	err := opt.Validate()
	if err != ErrDestAddressEmpty {
		t.Errorf("Expected ErrDestAddressEmpty, but got: %v", err)
	}
}
