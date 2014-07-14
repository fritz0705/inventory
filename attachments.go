package inventory

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

type AttachmentStore interface {
	Create() (a io.WriteCloser, id []byte, err error)
	Open(id []byte) (io.ReadCloser, error)
	Delete(id []byte) error
}

type FileAttachmentStore struct {
	Base string
}

func newAttachmentID() []byte {
	res := make([]byte, 16)
	_, err := rand.Read(res)
	if err != nil {
		panic(err)
	}
	return res
}

func (s *FileAttachmentStore) Create() (io.WriteCloser, []byte, error) {
	id := newAttachmentID()
	f, err := os.Create(filepath.Join(s.Base, hex.EncodeToString(id)))
	return f, id, err
}

func (s *FileAttachmentStore) Open(id []byte) (io.ReadCloser, error) {
	f, err := os.OpenFile(filepath.Join(s.Base, hex.EncodeToString(id)), os.O_RDWR, 0)
	return f, err
}

func (s *FileAttachmentStore) Delete(id []byte) error {
	return os.Remove(filepath.Join(s.Base, hex.EncodeToString(id)))
}

func (app *Application) AttachmentsHandler(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	key, err := hex.DecodeString(id)
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	object, err := app.AttachmentStore.Open(key)
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}
	defer object.Close()

	_, err = io.Copy(w, object)
	if err != nil {
		app.Error(w, err)
		return
	}
}

func (app *Application) PartUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.NotFoundHandler(w, r)
		return
	}

	id, _ := strconv.Atoi(path.Base(r.URL.Path))

	file, _, err := r.FormFile("file")
	if err != nil {
		app.Error(w, err)
		return
	}
	defer file.Close()

	object, key, err := app.AttachmentStore.Create()
	if err != nil {
		app.Error(w, err)
		return
	}
	defer object.Close()

	_, err = io.Copy(object, file)
	if err != nil {
		app.Error(w, err)
		return
	}

	attachment := Attachment{
		Key:    key,
		Name:   r.FormValue("name"),
		Type:   r.FormValue("type"),
		PartId: int64(id),
	}

	err = attachment.Save(app.DB)
	if err != nil {
		app.Error(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", id), http.StatusSeeOther)
}
