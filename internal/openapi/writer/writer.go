package writer

import (
	"context"

	"github.com/unkn0wn-root/resterm/internal/openapi"
	"github.com/unkn0wn-root/resterm/internal/restfile"
	"github.com/unkn0wn-root/resterm/internal/restwriter"
)

type FileWriter struct{}

func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

func (w *FileWriter) WriteDocument(
	ctx context.Context,
	doc *restfile.Document,
	destination string,
	opts openapi.WriterOptions,
) error {
	return restwriter.WriteDocument(ctx, doc, destination, restwriter.Options{
		OverwriteExisting: opts.OverwriteExisting,
		HeaderComment:     opts.HeaderComment,
	})
}
