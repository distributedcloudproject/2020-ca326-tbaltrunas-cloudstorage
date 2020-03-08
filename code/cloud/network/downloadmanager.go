package network

import (
	"cloud/utils"
	"fmt"
	"math/rand"
	"os"
	"sync"
)

const DOWNLOAD_WORKERS = 10

type DownloadEvent int

const (
	DownloadCompleted DownloadEvent = iota
	InfoRetrieved
)

type DownloadQueue struct {
	CloudPath       string
	LocalPath       string
	ChunkDownloaded []bool
	Completed       bool
	OnEvent         func(event DownloadEvent)
}

type DownloadManager struct {
	Cloud              *cloud
	queue              []*DownloadQueue
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

func (m *DownloadManager) QueueDownload(cloudPath, localPath string, OnComplete func(event DownloadEvent)) *DownloadQueue {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.queue) == 0 && m.downloading == 0 {
		m.completedFromQueue = 0
	}

	q := &DownloadQueue{
		CloudPath: cloudPath,
		LocalPath: localPath,
		OnEvent:   OnComplete,
	}
	m.queue = append(m.queue, q)

	if m.workers < DOWNLOAD_WORKERS {
		go m.createWorker()
		m.workers++
		m.downloading++
	}
	return q
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

		err := item.downloadFile(m.Cloud)
		if err != nil {
			fmt.Println(err)
			utils.GetLogger().Printf("Downloading file: %v", err)
		}
		m.mutex.Lock()
		m.downloading--
		m.completedFromQueue++
		m.mutex.Unlock()
		item.Completed = true
		if item.OnEvent != nil {
			item.OnEvent(DownloadCompleted)
		}
	}
}

// DownloadFile downloads the file from the cloud.

func (m *DownloadQueue) downloadFile(c *cloud) error {
	file, err := c.GetFile(m.CloudPath)
	if err != nil {
		return err
	}
	m.ChunkDownloaded = make([]bool, len(file.Chunks.Chunks))
	if m.OnEvent != nil {
		m.OnEvent(InfoRetrieved)
	}
	w, err := os.OpenFile(m.LocalPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer w.Close()
	dl := make([]int, len(file.Chunks.Chunks))
	for i := 0; i < len(dl); i++ {
		dl[i] = i
	}
	rand.Shuffle(len(dl), func(i, j int) { dl[i], dl[j] = dl[j], dl[i] })
	for i := range dl {
		chunk := file.Chunks.Chunks[dl[i]]
		content, err := c.GetChunk(m.CloudPath, chunk.ID)
		if err != nil {
			return err
		}
		m.ChunkDownloaded[dl[i]] = true
		_, err = w.Write(content)
		if err != nil {
			return err
		}
	}
	return nil
}
