package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/exercism/cli/api"
	"github.com/exercism/cli/config"
	"github.com/spf13/cobra"
)

// prepareCmd does necessary setup for Exercism and its tracks.
var prepareCmd = &cobra.Command{
	Use:     "prepare",
	Aliases: []string{"p"},
	Short:   "Prepare does setup for Exercism and its tracks.",
	Long: `Prepare downloads settings and dependencies for Exercism and the language tracks.

When called without any arguments, this downloads all the copy for the CLI so we
know what to say in all the various situations. It also provides an up-to-date list
of the API endpoints to use.

When called with a track ID, it will do specific setup for that track. This
might include downloading the files that the track maintainers have said are
necessary for the track in general. Any files that are only necessary for a specific
exercise will be downloaded along with the exercise.

To customize the CLI to suit your own preferences, use the configure command.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		track, err := cmd.Flags().GetString("track")
		if err != nil {
			return err
		}

		if track == "" {
			fmt.Println("prepare called")
			return nil
		}
		return prepareTrack(track)
	},
}

func prepareTrack(id string) error {
	usrCfg, err := config.NewUserConfig()
	if err != nil {
		return err
	}
	apiCfg, err := config.NewAPIConfig()
	if err != nil {
		return err
	}

	client, err := api.NewClient(usrCfg.Token, apiCfg.BaseURL)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/tracks/%s", apiCfg.BaseURL, id)

	req, err := client.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	payload := &prepareTrackPayload{}

	if err := json.NewDecoder(res.Body).Decode(payload); err != nil {
		return fmt.Errorf("unable to parse API response - %s", err)
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(payload.Error.Message)
	}

	cliCfg, err := config.NewCLIConfig()
	if err != nil {
		return err
	}

	t, ok := cliCfg.Tracks[id]
	if !ok {
		t = config.NewTrack(id)
	}
	if payload.Track.TestPattern != "" {
		t.IgnorePatterns = append(t.IgnorePatterns, payload.Track.TestPattern)
	}
	cliCfg.Tracks[id] = t

	return cliCfg.Write()
}

type prepareTrackPayload struct {
	Track struct {
		ID          string `json:"id"`
		Language    string `json:"language"`
		TestPattern string `json:"test_pattern"`
	} `json:"track"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func initPrepareCmd() {
	prepareCmd.Flags().StringP("track", "t", "", "the track you want to prepare")
}

func init() {
	RootCmd.AddCommand(prepareCmd)
	initPrepareCmd()
}
