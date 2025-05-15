package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"time"

	"github.com/disintegration/imaging" //nolint:depguard
)

func DownloadImage(url string, headers http.Header) ([]byte, error) {
	// Создаем контекст с таймаутом 10 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Создаем новый HTTP-запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Копируем заголовки из исходного запроса
	req.Header = headers

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Гарантируем закрытие тела ответа
	defer func() {
		_ = resp.Body.Close()
	}()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status code %d", resp.StatusCode)
	}

	// Читаем тело ответа
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ResizeImage(data []byte, width, height int) ([]byte, string, error) {
	// Декодируем изображение
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}

	resized := imaging.Resize(img, width, height, imaging.Lanczos)
	// Создаем буфер для сохранения результата
	var buf bytes.Buffer
	// Кодируем изображение в исходном формате
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resized, nil)
	case "png":
		err = png.Encode(&buf, resized)
	case "gif":
		err = gif.Encode(&buf, resized, nil)
	default:
		return nil, "", fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return nil, "", err
	}

	return buf.Bytes(), format, nil
}
