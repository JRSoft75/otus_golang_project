package image

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
)

func DownloadImage(url string, headers http.Header) ([]byte, error) {
	// Создаем новый HTTP-запрос
	req, err := http.NewRequest("GET", url, nil)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// Читаем тело ответа
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status code %d", resp.StatusCode)
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

	return buf.Bytes(), format, nil
}
