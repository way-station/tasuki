package shared

import "testing"

func TestParsePace(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{in: "3:28", want: 208},
		{in: "03:28", want: 208},
		{in: "12:05", want: 725},
		{in: "0:00", wantErr: true},
		{in: "3:60", wantErr: true},
		{in: "3:5", wantErr: true},
		{in: "100:00", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParsePace(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParsePace(%q) = %v, want an error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePace(%q): %v", tt.in, err)
			}
			if got.Seconds() != tt.want {
				t.Errorf("ParsePace(%q) = %d seconds, want %d", tt.in, got.Seconds(), tt.want)
			}
		})
	}
}

func TestPaceString(t *testing.T) {
	p, err := PaceFromSeconds(208)
	if err != nil {
		t.Fatal(err)
	}
	if got := p.String(); got != "3:28" {
		t.Errorf("String() = %q, want %q", got, "3:28")
	}
}

func TestTotalTime(t *testing.T) {
	got, err := ParseTotalTime("2:45:30")
	if err != nil {
		t.Fatal(err)
	}
	if got.Seconds() != 9930 {
		t.Errorf("Seconds() = %d, want 9930", got.Seconds())
	}
	if got.String() != "2:45:30" {
		t.Errorf("String() = %q, want %q", got.String(), "2:45:30")
	}
	for _, in := range []string{"0:00:00", "2:60:00", "45:30", ""} {
		if _, err := ParseTotalTime(in); err == nil {
			t.Errorf("ParseTotalTime(%q): want an error", in)
		}
	}
}
