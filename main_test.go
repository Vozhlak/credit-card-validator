package main

import (
	"io/fs"
	"os"
	"testing"
)

func TestExtractDigits(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"4000000000000002", "4000000000000002"},
		{"4000 0000 0000 0002", "4000000000000002"},
		{"4000-0000-0000-0002", "4000000000000002"},
		{"4000a000b000c0002", "40000000000002"},
		{"  -  ", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := extractDigits(tt.input)
		if got != tt.expected {
			t.Errorf("extractDigits(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestLoadBankData(t *testing.T) {
	content := `Lunar Bank,400000,400000
Mars Credit Union,500000,599999
`
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	banks, err := loadBankData(tmpFile)
	if err != nil {
		t.Fatalf("loadBankData failed: %v", err)
	}

	if len(banks) != 2 {
		t.Errorf("expected 2 banks, got %d", len(banks))
	}
	if banks[0].Name != "Lunar Bank" || banks[0].BinFrom != 400000 {
		t.Errorf("bank[0] = %+v", banks[0])
	}
	if banks[1].BinTo != 599999 {
		t.Errorf("bank[1] = %+v", banks[1])
	}
}

func TestExtractBIN(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"4000001234567890", 400000, false},
		{"123456", 123456, false},
		{"12345", 0, true}, // too short
	}

	for _, tt := range tests {
		got, err := extractBIN(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("extractBIN(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("extractBIN(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestIdentifyBank(t *testing.T) {
	banks := []Bank{
		{"Visa", 400000, 400000},
		{"Mastercard", 500000, 599999},
	}

	tests := []struct {
		bin      int
		expected string
	}{
		{400000, "Visa"},
		{550000, "Mastercard"},
		{999999, UnknownBank},
	}

	for _, tt := range tests {
		got := identifyBank(tt.bin, banks)
		if got != tt.expected {
			t.Errorf("identifyBank(%d) = %q, want %q", tt.bin, got, tt.expected)
		}
	}
}

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"79927398713", true},
		{"4000000000000002", true},
		{"4222222222222220", true},
		{"1234567890123456", false},
		{"123456s8v0A23456", false},
		{"0", true},
		{"1", false},
		{"", false},
	}

	for _, tt := range tests {
		got := validateLuhn(tt.input)
		if got != tt.expected {
			t.Errorf("validateLuhn(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

// Вспомогательная функция для тестов
func createTempFile(t *testing.T, content string) string {
	tmpFile := t.TempDir() + "/banks.txt"
	err := os.WriteFile(tmpFile, []byte(content), fs.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	return tmpFile
}
