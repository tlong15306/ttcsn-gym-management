package main

import "testing"

func TestMigrationCommandSafety(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		database string
		wantErr  bool
	}{
		{"up dev", "up", "postgres://localhost/ttcsn_gym_dev", false},
		{"status dev", "status", "postgres://localhost/ttcsn_gym_dev", false},
		{"version dev", "version", "postgres://localhost/ttcsn_gym_dev", false},
		{"down test", "down", "postgres://localhost/ttcsn_gym_test", false},
		{"reset test", "reset", "postgres://localhost/ttcsn_gym_test", false},
		{"down dev", "down", "postgres://localhost/ttcsn_gym_dev", true},
		{"reset unrelated", "reset", "postgres://localhost/master-golang", true},
		{"test name prefix", "down", "postgres://localhost/ttcsn_gym_test_copy", true},
		{"missing database", "reset", "postgres://localhost", true},
		{"invalid URL", "down", "postgres://%invalid", true},
		{"unknown command", "create", "postgres://localhost/ttcsn_gym_test", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateCommand(tt.command, tt.database); (err != nil) != tt.wantErr {
				t.Fatalf("validateCommand() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}
