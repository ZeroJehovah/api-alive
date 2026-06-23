package alive

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const maxHumanErrorChars = 30

func PrintHuman(w io.Writer, res Result) {
	PrintHumanAligned(w, res, len(res.Model))
}

func PrintHumanAligned(w io.Writer, res Result, modelWidth int) {
	if modelWidth < len(res.Model) {
		modelWidth = len(res.Model)
	}
	if res.Success {
		fmt.Fprintf(w, "%-*s  %8dms  attempts=%d  %s\n", modelWidth, res.Model, res.DurationMS, res.Attempts, resultStatus(res))
		return
	}
	fmt.Fprintf(w, "%-*s  %8dms  attempts=%d  error=%q  %s\n", modelWidth, res.Model, res.DurationMS, res.Attempts, humanError(res.Error), resultStatus(res))
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

func humanError(err string) string {
	err = strings.Join(strings.Fields(err), " ")
	return truncateRunes(err, maxHumanErrorChars)
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if max <= 0 || len(runes) <= max {
		return s
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

func PrintJSONLine(w io.Writer, res Result) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
