package command

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Xz ...
func Xz(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Если аутяга ничего не вложил для сжатия
	if len(m.Attachments) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Кто прочитал тот здохнет")
		return
	}

	// Собираем из вложенного урл, имя файлы и размер
	fileURL := m.Attachments[0].URL
	fileName := m.Attachments[0].Filename
	fileSize := m.Attachments[0].Size

	// Получаем файл
	response, err := http.Get(fileURL)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Иди нахуй, короче")
		log.Println(err)
		return
	}
	defer response.Body.Close()

	// Сжимаем файл
	file := new(bytes.Buffer)
	file.ReadFrom(response.Body)

	out, err := xz(file.Bytes())
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Если вкратце, то иди нахуй")
		log.Println(err)
		return
	}

	// Создаём из буффера байтиков ридер для отправки
	outReader := strings.NewReader(string(out))

	// Вычисляем различие в размере
	diffSize := len(string(out)) - fileSize

	// Корректно строим сообщения на основе diffSize
	message := strings.Builder{}
	if diffSize > 0 {
		fmt.Fprintf(&message, "Случился пиздец +%d", diffSize)
	} else if diffSize >= -100 {
		fmt.Fprintf(&message, "Хуёво %d", diffSize)
	} else {
		fmt.Fprintf(&message, "Норм %d", diffSize)
	}

	// Отправка файла сo смешным сообщением
	s.ChannelFileSendWithMessage(m.ChannelID, message.String(), fileName+".xz", outReader)

}

// xz принимает байтики, сжимает и возвращает, также, байтики
func xz(in []byte) (out []byte, err error) {
	cmd := exec.Command("xz", "-e9vfT0", "--memlimit=1600MiB", "--stdout")

	cmd.Stdin = strings.NewReader(string(in))
	var o bytes.Buffer

	cmd.Stdout = &o

	err = cmd.Run()
	if err != nil {
		return
	}

	out = o.Bytes()
	return
}
