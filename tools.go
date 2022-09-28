package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// characters for generating random string
const randStringSource = "abcdefhijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

// Tools type used for instantiation.
// variables of this type will have access to methods with a *Tools reciever
type Tools struct {
	MaxFileSize      int64
	AllowedFileTypes []string
}

// RandomString returns a string of random characters of length n, sourced from randStringSource
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randStringSource)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

// UploadedFile Saves information about an uploaded file
type UploadedFile struct {
	NewFileName string
	OldFileName string
	FileSize    int64
}

func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024 // 1GB
	}
	err := r.ParseMultipartForm(t.MaxFileSize)
	if err != nil {
		return nil, errors.New("error: File size is too large")
	}
	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(file []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {

				}
				defer func(infile multipart.File) {
					err := infile.Close()
					if err != nil {
						fmt.Println(err)
					}
				}(infile)
				buff := make([]byte, 512) // why 512 bytes ? see http://golang.org/pkg/net/http/#DetectContentType
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				// DANGER: http.DetectContentType will only use the first 512 bytes to sniff the content type.
				allowed := false
				fileType := http.DetectContentType(buff)
				if len(t.AllowedFileTypes) > 0 {
					for _, t := range t.AllowedFileTypes {
						if t == fileType {
							allowed = true
						}
					}
				} else {
					allowed = true
				}
				if !allowed {
					return nil, errors.New("error: File type is not allowed")
				}
				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(20), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				var outfile *os.File
				defer func(outfile *os.File) {
					err := outfile.Close()
					if err != nil {
						fmt.Println(err)
					}
				}(outfile)
				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				}
				uploadedFiles = append(uploadedFiles, &uploadedFile)
				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, nil
}
