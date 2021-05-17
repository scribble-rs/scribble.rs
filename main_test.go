package main

import (
	"flag"
	"os"
	"testing"
)

func TestDeterminePort(t *testing.T) {
	type testCase struct {
		name string

		portFlag int
		portEnv  string

		expectedPort int
	}

	cases := []*testCase{
		{
			name:         "no input",
			portFlag:     -1,
			expectedPort: defaultPort,
		},
		{
			name:         "invalid port flag",
			portFlag:     -1,
			expectedPort: defaultPort,
		},
		{
			name:         "invalid port flag",
			portFlag:     -2,
			expectedPort: defaultPort,
		},
		{
			name:         "invalid port flag",
			portFlag:     -70000,
			expectedPort: defaultPort,
		},
		{
			name:         "invalid port flag",
			portFlag:     65537,
			expectedPort: defaultPort,
		},
		{
			name:         "invalid port flag",
			portFlag:     70000,
			expectedPort: defaultPort,
		},
		{
			name:         "port flag over env variable",
			portFlag:     50000,
			portEnv:      "1234",
			expectedPort: 50000,
		},
		{
			name:         "valid env variable",
			portFlag:     -1,
			portEnv:      "1234",
			expectedPort: 1234,
		},
		{
			name:         "valid env variable with space",
			portFlag:     -1,
			portEnv:      " 1234 ",
			expectedPort: 1234,
		},
		{
			name:         "invalid env variable with space",
			portFlag:     -1,
			portEnv:      " -1234 ",
			expectedPort: defaultPort,
		},
		{
			name:         "unparsable env variable",
			portFlag:     -1,
			portEnv:      "hello there",
			expectedPort: defaultPort,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, arg := range flag.CommandLine.Args() {
				t.Logf("Removing arg %s", arg)
			}
			os.Setenv("PORT", c.portEnv)

			port := determinePort(c.portFlag)
			if port != c.expectedPort {
				t.Errorf("Port was %d instead of %d", port, c.expectedPort)
			}
		})
	}
}
