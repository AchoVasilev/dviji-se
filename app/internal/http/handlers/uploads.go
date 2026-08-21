package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"server/internal/http/middleware"
	"server/util/httputils"
	"slices"
	"strings"
	"time"
)

// uploadTime is the budget for an upload request. The ordinary handler budget
// is ten seconds, which a 20 MiB file on a phone connection cannot meet.
var uploadTime = 5 * time.Minute

// uploadParseMemory is how much of a multipart form is held in memory; the
// rest spills to a temporary file that the server removes when the request
// ends. Kept well below the size limit so a large upload cannot be turned
// into 20 MiB of resident memory per request.
const uploadParseMemory = 4 << 20 // 4 MiB

// allowedImageTypes are the formats accepted for images. Detection is by
// content rather than by the extension or the browser's Content-Type, both of
// which the client chooses.
var allowedImageTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
}

// readUpload parses the multipart form and hands back the single uploaded
// file. It answers the request itself and reports false when anything is
// wrong, so the callers stay linear.
func readUpload(ctx context.Context, w http.ResponseWriter, r *http.Request) (multipart.File, *multipart.FileHeader, bool) {
	// The whole point of a large limit is that the transfer takes a while, and
	// the server's read timeout is set for ordinary requests. Extending the
	// deadline here keeps a slow upload from being cut off mid-body.
	if err := http.NewResponseController(w).SetReadDeadline(time.Now().Add(uploadTime)); err != nil {
		slog.WarnContext(ctx, "Could not extend the upload read deadline", "error", err)
	}

	if err := r.ParseMultipartForm(uploadParseMemory); err != nil {
		// MaxBytesReader stops the body at the limit, and that arrives here as
		// a parse failure. Reporting it as a bad form would tell the author
		// nothing about what to do differently.
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			httputils.SendErrorResponse(ctx, w, tooLargeMessage(), http.StatusRequestEntityTooLarge)
			return nil, nil, false
		}

		slog.WarnContext(ctx, "Could not parse the upload form", "error", err)
		httputils.SendBadRequestResponse(ctx, w, "Файлът не можа да бъде прочетен")

		return nil, nil, false
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httputils.SendBadRequestResponse(ctx, w, "Няма прикачен файл")
		return nil, nil, false
	}

	// The body limit allows for multipart overhead, so a file can still be
	// over the per-file limit while the request as a whole fits.
	if header.Size > middleware.MaxUploadFileBytes {
		file.Close()
		httputils.SendErrorResponse(ctx, w, tooLargeMessage(), http.StatusRequestEntityTooLarge)

		return nil, nil, false
	}

	return file, header, true
}

// requireImage rejects anything that is not one of the accepted image formats.
// The file is rewound afterwards so the caller can still upload it.
func requireImage(file multipart.File) error {
	// DetectContentType looks at the first 512 bytes at most.
	head := make([]byte, 512)

	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.New("Файлът не можа да бъде прочетен")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errors.New("Файлът не можа да бъде прочетен")
	}

	// DetectContentType returns "image/png" but also "text/plain; charset=..."
	// for anything it does not recognise, so the parameters are cut off first.
	contentType, _, _ := strings.Cut(http.DetectContentType(head[:n]), ";")
	if !slices.Contains(allowedImageTypes, contentType) {
		return errors.New("Позволени са само изображения (JPEG, PNG, GIF, WebP)")
	}

	return nil
}

func tooLargeMessage() string {
	return fmt.Sprintf("Файлът е твърде голям. Максимум %d MB.", middleware.MaxUploadFileMB())
}
