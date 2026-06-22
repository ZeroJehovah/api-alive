package alive

import (
	"encoding/json"
	"fmt"
	"io"
)

func PrintHuman(w io.Writer, res Result) {
	status := "success"
	if !res.Success {
		status = "failed"
	}
	fmt.Fprintf(w, "model=%s provider=%s result=%s duration=%dms\n", res.Model, res.Provider, status, res.DurationMS)
	if !res.Success {
		if res.ExitCode != nil {
			fmt.Fprintf(w, "exit_code=%d\n", *res.ExitCode)
		}
		if res.Error != "" {
			fmt.Fprintf(w, "error=%s\n", res.Error)
		}
	}
	fmt.Fprintf(w, "prompt=%q expected=%q\n", res.Prompt, res.Expected)
	if res.Output != "" {
		fmt.Fprintf(w, "output:\n%s\n", res.Output)
	}
	fmt.Fprintln(w, "---")
}

func PrintJSONLine(w io.Writer, res Result) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
