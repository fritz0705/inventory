package inventory

import (
	"crypto/rand"
	"database/sql"
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

func (app *Application) PartUploadDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.NotFoundHandler(w, r)
		return
	}

	tx := app.DB.MustBegin()
	defer tx.Rollback()

	id := path.Base(r.URL.Path)
	attachment := new(Attachment)
	err := tx.Get(attachment, `SELECT * FROM 'attachment' WHERE "id" = ?`, id)
	if err != nil {
		app.Error(w, err)
		app.NotFoundHandler(w, r)
		return
	}

	_, err = tx.Exec(`DELETE FROM 'attachment' WHERE "id" = ?`, id)
	if err != nil {
		app.Error(w, err)
		return
	}

	tx.Commit()

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%v", id), http.StatusSeeOther)
}

func (app *Application) PartUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.NotFoundHandler(w, r)
		return
	}

	tx := app.DB.MustBegin()
	defer tx.Rollback()

	id, _ := strconv.Atoi(path.Base(r.URL.Path))

	part := new(Part)
	err := tx.Get(part, `SELECT * FROM 'part' WHERE "id" = ?`, id)
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	file, header, err := r.FormFile("file")
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
		Name:   header.Filename,
		Type:   header.Header.Get("Content-Type"),
		PartId: int64(id),
	}

	err = attachment.Save(tx)
	if err != nil {
		app.Error(w, err)
		return
	}

	if !part.ImageId.Valid {
		print("Fehler!")
		part.ImageId = sql.NullInt64{attachment.Id, true}
		err = part.Save(tx)
		if err != nil {
			app.Error(w, err)
			return
		}
	}

	tx.Commit()

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", id), http.StatusSeeOther)
}
