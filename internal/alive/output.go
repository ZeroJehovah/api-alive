package alive

import (
	"encoding/json"
	"fmt"
	"io"
)

func PrintHuman(w io.Writer, res Result) {
	PrintHumanAligned(w, res, len(res.Model))
}

func PrintHumanAligned(w io.Writer, res Result, modelWidth int) {
	if modelWidth < len(res.Model) {
		modelWidth = len(res.Model)
	}
	fmt.Fprintf(w, "%-*s  %8dms  %-7s  attempts=%d\n", modelWidth, res.Model, res.DurationMS, resultStatus(res), res.Attempts)
}

func ModelColumnWidth(models []string) int {
	width := 0
	for _, model := range models {
		if len(model) > width {
			width = len(model)
		}
	}
	return width
}

func resultStatus(res Result) string {
	if res.Success {
		return "success"
	}
	return "failed"
}

func PrintJSONLine(w io.Writer, res Result) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
