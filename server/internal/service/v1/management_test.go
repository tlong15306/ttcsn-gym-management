package v1service

import (
	"bytes"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/member"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/trainer"
	"image"
	"image/png"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWorkingHoursBoundaries(t *testing.T) {
	cases := []struct {
		name  string
		hours []trainer.WorkingHour
		valid bool
	}{
		{"adjacent", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "08:00", EndTime: "09:00"}, {DayOfWeek: 1, StartTime: "09:00", EndTime: "10:00"}}, true},
		{"overlap", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "08:00", EndTime: "09:00"}, {DayOfWeek: 1, StartTime: "08:59", EndTime: "10:00"}}, false},
		{"different days", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "08:00", EndTime: "09:00"}, {DayOfWeek: 2, StartTime: "08:00", EndTime: "09:00"}}, true},
		{"invalid day", []trainer.WorkingHour{{DayOfWeek: 0, StartTime: "08:00", EndTime: "09:00"}}, false},
		{"invalid time", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "24:00", EndTime: "25:00"}}, false},
		{"empty interval", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "09:00", EndTime: "09:00"}}, false},
		{"empty schedule", []trainer.WorkingHour{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validHours(tc.hours); got != tc.valid {
				t.Fatalf("validHours=%v, want %v", got, tc.valid)
			}
		})
	}
}
func TestMemberDatesAndContactValidation(t *testing.T) {
	in := member.Input{MemberCode: "M001", JoinDate: "2026-10-01", MembershipExpiry: "2026-11-01"}
	if !validMember(in, true) {
		t.Fatal("valid dates rejected")
	}
	in.MembershipExpiry = "2026-09-01"
	if validMember(in, true) {
		t.Fatal("expiry before join accepted")
	}
	in.MembershipExpiry = "2026-11-31"
	if validMember(in, true) {
		t.Fatal("invalid day accepted")
	}
	if validContact("Name", "foo@@example.com", "") || validContact("Name", "Name <foo@example.com>", "") {
		t.Fatal("invalid email accepted")
	}
}
func TestCertificateContentValidation(t *testing.T) {
	var buf bytes.Buffer
	png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	if ext, err := certificateExtension(buf.Bytes()); err != nil || ext != ".png" {
		t.Fatal("valid PNG rejected", err)
	}
	for _, content := range [][]byte{[]byte("<html>fake.pdf</html>"), []byte("%PDF-1.7 truncated"), []byte("\x89PNG\r\n\x1a\nmalformed")} {
		if _, err := certificateExtension(content); err == nil {
			t.Fatal("malformed certificate accepted")
		}
	}
}

func TestPasswordCharacterAndByteLimits(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{{"1234567", false}, {"12345678", true}, {"ắắắ", false}, {"ắắắắắắắắ", true}, {strings.Repeat("a", 72), true}, {strings.Repeat("a", 73), false}, {strings.Repeat("ắ", 25), false}} {
		if got := validPassword(tc.value); got != tc.valid {
			t.Fatalf("password chars/bytes %d/%d valid=%v want=%v", utf8.RuneCountInString(tc.value), len(tc.value), got, tc.valid)
		}
	}
}
