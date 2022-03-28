package command

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// OptimizeCommand ...
func OptimizeCommand(s *discordgo.Session, m *discordgo.MessageCreate, iteration int) {
	cErr := CuteErr{s, m}

	// Если аутяга ничего не вложил для сжатия
	if len(m.Attachments) == 0 {
		cErr.WrongCommand()
		return
	}

	// Собираем из вложенного урл, имя файлы и размер
	fileURL := m.Attachments[0].URL
	fileName := m.Attachments[0].Filename
	fileSize := m.Attachments[0].Size
	tmpfileName := "tmp" + fileName

	// Получаем файл
	response, err := http.Get(fileURL)
	if err != nil {
		cErr.InsideError(err)
		return
	}
	defer response.Body.Close()

	file := new(bytes.Buffer)
	file.ReadFrom(response.Body)

	contentType := http.DetectContentType(file.Bytes())

	tmpFile, err := os.Create(fileName)
	if err != nil {
		cErr.InsideError(err)
		return
	}

	_, _ = io.Copy(tmpFile, strings.NewReader(file.String()))

	if strings.HasPrefix(contentType, "text/html") {
		contentType = "text/html"
	}

	fmt.Println(contentType)

	var cmds []*exec.Cmd
	switch contentType {
	case "image/png":
		cmds = append(cmds, exec.Command("pngquant", "--strip", "--quality=85-95", "--speed=1", "--ext=.png", "--force", fileName),
			exec.Command("pngout", "-y", "-r", "-d0", "-mincodes0", "-kacTL,fcTL,fdAT", fileName),
			exec.Command("optipng", "-zw32k", "-o7", "-quiet", "-strip all", fileName),
			exec.Command("leanify", "-i", "30", "-q", fileName),
			exec.Command("advpng", "-z", "-q", "-4", "-i", "30", fileName),
			exec.Command("ECT", "--allfilters", "--mt-deflate", "-strip", "--strict", "-quiet", "-9", fileName),
		)
	case "image/jpeg":
		cmds = append(cmds, exec.Command("guetzli", "--quality", "90", fileName, tmpfileName),
			exec.Command("jhead", "-q", "-autorot", "-purejpg", "-di", "-dx", "-dt", "-zt", fileName),
			exec.Command("leanify", "-i", "30", "-q", fileName),
			// exec.Command("magick", "convert", fileName, "-quiet", "-interlace", "Plane", "-define", "jpeg:optimize-coding=true", "-strip", tmpfileName),
			exec.Command("jpegoptim", "-o", "-q", "--all-progressive", "--strip-all", fileName),
			exec.Command("ECT", "--allfilters", "--mt-deflate", "-strip", "--strict", "-quiet", "-9", fileName),
		)
	case "image/gif":
		// exec.Command("magick", "convert", fileName, "-quiet", "-set", "dispose", "background", "-layers", "optimize", "-compress", "LZW", "-strip", tmpfileName)
		cmds = append(cmds, exec.Command("gifsicle", "-w", "-j", "--no-conserve-memory", "-o", tmpfileName, "--no-comments", "--no-extensions", "--no-names", "--lossy=85", "-O3", fileName))
	case "text/html":
		cmds = append(cmds, exec.Command("tidy", "-m", "--wrap", "0", "--bare", "yes", "--clean", "yes", "--indent", "0", "--join-classes", "yes", "--omit-optional-tags", "yes", "--tidy-mark", "no", "--quiet", "yes", fileName),
			exec.Command("leanify", "-i", "30", "-q", fileName),
		)
	case "video/webm":
		cmds = append(cmds, exec.Command("ffmpeg", "-i", fileName, "-map", "0", "-crf", "28", "-vbr", "4", "-b:a", "128k", "-preset", "veryslow", tmpfileName),
			exec.Command("mkclean", "--optimize", "--unsafe", "--quiet", fileName, tmpfileName),
		)
	case "video/mp4":
		cmds = append(cmds, exec.Command("ffmpeg", "-i", fileName, "-map", "0", "-crf", "28", "-vbr", "4", "-b:a", "128k", "-preset", "veryslow", tmpfileName),
			exec.Command("mp4file", "--optimize", "-q", fileName),
		)
	case "application/pdf":
		cmds = append(cmds, exec.Command("mutool", "clean", "-ggg", "-z", fileName, tmpfileName))
		// exec.Command("gs", "-dColorImageDownsampleType=/Bicubic", "-dGrayImageDownsampleType=/Bicubic", "-dMonoImageDownsampleType=/Bicubic", "-dOptimize=true", "-dConvertCMYKImagesToRGB=true", "-dColorConversionStrategy=/sRGB", "-dPrinted=false", "-q", "-dBATCH", "-dNOPAUSE", "-dSAFER", "-dDELAYSAFER", "-dNOPROMPT", "-sDEVICE=pdfwrite", "-dDetectDuplicateImages=true", "-dAutoRotatePages=/None", "-dCompatibilityLevel=1.4", "-dPDFSETTINGS=/ebook", "-dDownsampleColorImages=true", "-dColorImageResolution=150", "-dDownsampleGrayImages=true", "-dGrayImageResolution=150", "-dDownsampleMonoImages=true", "-dMonoImageResolution=150", "-sOutputFile=\""+tmpfileName+"\"", fileName)
	case "image/webp":
		cmds = append(cmds, exec.Command("dwebp", "-mt", "-o", tmpfileName, fileName),
			exec.Command("cwebp", "-mt", "-quiet", "-lossless", "-m", "6", fileName, "-o", tmpfileName),
		)
	case "application/octet-stream":
		if strings.HasSuffix(fileName, ".exe") {
			cmds = append(cmds, exec.Command("strip", "--strip-all", "-o", tmpfileName, fileName),
				exec.Command("upx", "--no-backup", "--force", "-9", "--best", "--lzma", "--ultra-brute", "--crp-ms=999999", fileName),
				exec.Command("leanify", "-i", "30", "-q", fileName),
			)
		} else if strings.HasSuffix(fileName, ".dll") {
			cmds = append(cmds, exec.Command("strip", "--strip-all", "-o", tmpfileName, fileName),
				exec.Command("upx", "--no-backup", "--force", "-9", "--best", "--lzma", "--ultra-brute", "--crp-ms=999999", fileName),
			)
		}

	default:
		cErr.WrongCommand()
		os.Remove(fileName)
		return
	}

	var percent, count int = 0, 0
	msg, _ := s.ChannelMessageSend(m.ChannelID, "Понял")

	for j := 0; j < iteration; j++ {
		for _, cmd := range cmds {

			executeA := func(cmd exec.Cmd) (err error) {
				err = cmd.Run()
				if err != nil {
					return err
				}
				return nil
			}

			err = executeA(*cmd)

			if err != nil && contentType != "text/html" {
				fmt.Printf("%s: %+v", "External command", err)
				s.ChannelMessageEdit(m.ChannelID, msg.ID, "Лень")
				continue
			}

			_, err := os.Stat(tmpfileName)
			if !os.IsNotExist(err) {
				os.Remove(fileName)
				os.Rename(tmpfileName, fileName)
			}

			out, err := ioutil.ReadFile(tmpFile.Name())
			if err != nil {
				cErr.InsideError(err)
				return
			}

			count++
			progressMessage := strings.Builder{}
			percent = PercentageChange(fileSize, len(string(out)))
			fmt.Fprintf(&progressMessage, "%s (%d/%d) %d/%d | %d(%d%%)", cmd.Args[0], count, len(cmds)*iteration, fileSize, len(string(out)), len(string(out))-fileSize, percent)

			s.ChannelMessageEdit(m.ChannelID, msg.ID, progressMessage.String())

			fmt.Println(progressMessage.String())
		}
	}

	s.ChannelMessageDelete(msg.ChannelID, msg.ID)

	out, err := ioutil.ReadFile(tmpFile.Name())

	os.Remove(fileName)
	os.Remove(tmpfileName)

	message := strings.Builder{}
	fmt.Fprintf(&message, "%s %d/%d | %d(%d%%)", "LOLWAT?", fileSize, len(string(out)), len(string(out))-fileSize, PercentageChange(fileSize, len(string(out))))

	outReader := strings.NewReader(string(out))
	s.ChannelFileSendWithMessage(m.ChannelID, message.String(), fileName, outReader)
}

// PercentageChange ...
func PercentageChange(old, new int) (delta int) {
	diff := float64(new - old)
	delta = int((diff / float64(old)) * 100)
	return
}
