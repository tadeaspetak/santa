package app

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/go-pdf/fpdf"
	"github.com/k3a/html2text"
	"github.com/tadeaspetak/santa/internal/data"
)

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

	message := removeSymbols(html2text.HTML2Text(fmt.Sprintf("%s\n\n%s", subject, body)))
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

func Send(mlr mailer, pairs []participantPair, template data.Template, isDebug bool, alwaysSendTo string) error {
	batchDate := time.Now().Local().Format("20060102-150405")

	pdf := fpdf.New("P", "mm", "A4", "")
	// Set all margins to 10 **except** the bottom one. The bottom one is
	// set to 17 for the convenience of calculation. That way, each page has
	// 297 - 10 - 17 = 270mm of height at its disposal.
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 17)

	for _, pair := range pairs {
		// prefer the email provided via the `alwaysSendTo` flag for testing purposes
		recipient := alwaysSendTo
		if recipient == "" {
			recipient = pair.giver.Email
		}

		replacer := strings.NewReplacer("%{recipientSalutation}", pair.recipient.Salutation)
		subject := replacer.Replace(template.Subject)
		body := replacer.Replace(fmt.Sprintf(`<html><body>%s</body></html>`, template.Body))

		// write a history batch file
		err := os.WriteFile(
			fmt.Sprintf("santa-batch-%s-%s.txt", batchDate, pair.giver.Email),
			[]byte(fmt.Sprintf("Sent to: %s\nSubject: %s\nBody: %s", pair.giver.Email, subject, body)),
			0644,
		)
		if err != nil {
			return fmt.Errorf("unable to write a history batch file %v", err)
		}

		addPDFPage(pdf, pair.giver.Email, subject, body)

		// don't send anything when debugging
		if isDebug {
			fmt.Printf("DEBUG: Would be sending an email to %s (batch file generated).\n", recipient)
			continue
		}

		err = mlr.send(
			template.Sender,
			subject,
			body,
			recipient,
			pair.giver.Email, // even when a fixed recipient is present, set the reply-to to the actual giver's email to make debugging easy
		)

		if err != nil {
			// Note: It's questionable what the best course of action is here. Is it better to continue
			// with the current batch or return an error even though some emails may already have been sent out?
			return fmt.Errorf("could not send email to %v: %v", recipient, err)
		}

		fmt.Printf("Email to %s sent successfully\n", recipient)

	}

	err := pdf.OutputFileAndClose(fmt.Sprintf("santa-batch-%s.pdf", batchDate))
	if err != nil {
		log.Fatalf("Could not generate PDF: %v", err)
	}

	return nil
}
