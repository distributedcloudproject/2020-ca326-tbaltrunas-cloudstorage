package network

import (
	"fmt"
	"os"
	"sync"
)

const DOWNLOAD_WORKERS = 10

type DownloadQueue struct {
	CloudPath  string
	LocalPath  string
	OnComplete func()
}

type DownloadManager struct {
	Cloud              *cloud
	queue              []DownloadQueue
	completedFromQueue int
	downloading        int

	mutex   sync.Mutex
	workers int
}

func (m *DownloadManager) Queued() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.queue)
}

func (m *DownloadManager) Downloading() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.downloading
}

func (m *DownloadManager) Completed() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.completedFromQueue
}

func (m *DownloadManager) QueueDownload(cloudPath, localPath string, OnComplete func()) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.queue) == 0 && m.downloading == 0 {
		m.completedFromQueue = 0
	}

	m.queue = append(m.queue, DownloadQueue{
		CloudPath:  cloudPath,
		LocalPath:  localPath,
		OnComplete: OnComplete,
	})

	if m.workers < DOWNLOAD_WORKERS {
		go m.createWorker()
		m.workers++
		m.downloading++
	}
}

func (m *DownloadManager) createWorker() {
	defer func() {
		m.mutex.Lock()
		m.workers--
		m.mutex.Unlock()
	}()

	for len(m.queue) != 0 {
		m.mutex.Lock()
		if len(m.queue) == 0 {
			m.mutex.Unlock()
			return
		}

		item := m.queue[0]
		m.queue = append(m.queue[1:])
		m.mutex.Unlock()

		m.downloadFile(item.CloudPath, item.LocalPath)
		m.mutex.Lock()
		m.downloading--
		m.completedFromQueue++
		m.mutex.Unlock()
		item.OnComplete()
	}
}

// DownloadFile downloads the file from the cloud.
func (m *DownloadManager) downloadFile(cloudPath string, localPath string) error {
	fmt.Println("DOWNLOAD FILE: ", cloudPath, localPath)
	file, err := m.Cloud.GetFile(cloudPath)
	if err != nil {
		return err
	}
	w, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer w.Close()
	for _, chunk := range file.Chunks.Chunks {
		content, err := m.Cloud.GetChunk(cloudPath, chunk.ID)
		if err != nil {
			return err
		}
		_, err = w.Write(content)
		if err != nil {
			return err
		}
	}
	return nil
}
