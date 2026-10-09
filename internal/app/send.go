package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/go-pdf/fpdf"
	"github.com/k3a/html2text"
	"github.com/tadeaspetak/santa/internal/data"
)

func mapOver[K any, V any](slice []K, predicate func(v K) V) []V {
	result := make([]V, len(slice))
	for i, v := range slice {
		result[i] = predicate(v)
	}
	return result
}

// removeSymbols removes symbols because fpdf fails e.g. on emojis.
//
// In fpdf, character widths are capped at 256*256 (https://github.com/jung-kurt/gofpdf/blob/514e371ce761f71cf004bf0da3246824310b2e4f/utf8fontfile.go#L838).
// See also https://github.com/jung-kurt/gofpdf/issues/255.
// Based on the above, I tried r < 65536, but passing a snowman ⛄ (9924) still breaks, so let's just skip symbols for now 🤷‍♀️
func removeSymbols(str string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if !unicode.IsSymbol(r) {
			return r
		}

		return -1
	}, str))
}

func addPDFPage(pdf *fpdf.Fpdf, emailRecipient, subject, body string) {
	multiLineHeight := 4
	availablePageHeight := 270

	cleanSubject := removeSymbols(subject)
	cleanBody := removeSymbols(HTMLToText(body))
	message := fmt.Sprintf("%s\n\n%s", cleanSubject, cleanBody)
	pdf.SetFont("Arial", "", 12)
	lines := pdf.SplitText(message, 190)

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(
		0,
		float64(availablePageHeight-(len(lines)*multiLineHeight)),
		emailRecipient,
		"",
		1,
		"CM",
		false,
		0,
		"",
	)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, float64(multiLineHeight), message, "", "L", false)
}

// Mailer sends emails, see the `mailer` package for the implementations.
type Mailer interface {
	Send(sender, subject, body, recipient, replyTo string) error
	// Verify checks that the mailer can log in without sending anything.
	Verify(ctx context.Context) error
}

type SendOpts struct {
	// To, when set, receives all the emails instead of the participants (for testing purposes).
	To string
	// IsDryRun generates everything as usual but doesn't send any emails.
	IsDryRun bool
	// ShouldPrintPdf generates a printable PDF with all the emails.
	ShouldPrintPdf bool
	// OutputDir is where the batch files (and the PDF) are written. Created if it doesn't exist.
	OutputDir string
	// Out receives the progress messages. Defaults to stdout.
	Out io.Writer
}

// Render builds the subject and the HTML body of an email for the given recipients.
func Render(template data.Template, recipients []data.Person) (subject string, body string) {
	separator := template.RecipientsSeparator
	if separator == "" {
		separator = " and "
	}
	salutations := mapOver(recipients, func(v data.Person) string { return v.Salutation })
	replacer := strings.NewReplacer("%{recipientSalutation}", strings.Join(salutations, separator))

	subject = replacer.Replace(template.Subject)
	body = replacer.Replace(fmt.Sprintf(`<html><body>%s</body></html>`, template.Body))
	return subject, body
}

// HTMLToText converts an email body to plain text, e.g. for printing it in the terminal.
func HTMLToText(html string) string {
	return html2text.HTML2Text(html)
}

func Send(
	mlr Mailer,
	pairs []giverWithRecipients,
	template data.Template,
	opts SendOpts,
) error {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}

	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("unable to create the output directory: %w", err)
	}

	var pdf *fpdf.Fpdf
	if opts.ShouldPrintPdf {
		pdf = fpdf.New("P", "mm", "A4", "")
		// Set all margins to 10 **except** the bottom one. The bottom one is
		// set to 17 for the convenience of calculation. That way, each page has
		// 297 - 10 - 17 = 270mm of height at its disposal.
		pdf.SetMargins(10, 10, 10)
		pdf.SetAutoPageBreak(true, 17)
		pdf.AddPage()
	}

	for i, pair := range pairs {
		// prefer the email provided via the `To` option for testing purposes
		recipient := opts.To
		if recipient == "" {
			recipient = pair.giver.Email
		}

		subject, body := Render(template, pair.recipients)

		// write a batch file per giver, useful for debugging and fixing typos in emails
		err := os.WriteFile(
			filepath.Join(opts.OutputDir, pair.giver.Email+".txt"),
			[]byte(fmt.Sprintf("Sent to: %s\nSubject: %s\nBody: %s", pair.giver.Email, subject, body)),
			0644,
		)
		if err != nil {
			return fmt.Errorf("unable to write a batch file: %w", err)
		}

		if opts.ShouldPrintPdf {
			addPDFPage(pdf, pair.giver.Email, subject, body)
		}

		// don't send anything on dry runs
		if opts.IsDryRun {
			fmt.Fprintf(out, "  - %s (not sent, dry run)\n", pair.giver.Email)
			continue
		}

		err = mlr.Send(
			template.From(),
			subject,
			body,
			recipient,
			pair.giver.Email, // even when a fixed email recipient is present, set the reply-to to the actual giver's email to make debugging easy
		)

		if err != nil {
			// Note: It's questionable what the best course of action is here. Is it better to continue
			// with the current batch or return an error even though some emails may already have been sent out?
			return fmt.Errorf("could not send email to %v (%d of %d sent before this one): %w", recipient, i, len(pairs), err)
		}

		if recipient == pair.giver.Email {
			fmt.Fprintf(out, "  ✓ %s\n", recipient)
		} else {
			fmt.Fprintf(out, "  ✓ %s (instead of %s)\n", recipient, pair.giver.Email)
		}
	}

	if opts.ShouldPrintPdf {
		pdf.AddPage()
		err := pdf.OutputFileAndClose(filepath.Join(opts.OutputDir, "emails.pdf"))
		if err != nil {
			return fmt.Errorf("could not generate PDF: %v", err)
		}
	}

	return nil
}
