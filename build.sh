#!/bin/bash
VER=$(git describe --tags)
> sha256-checksums

app="youtubeuploader_linux_armv7"
echo "Generating $app ..."
(env GOOS=linux GOARCH=arm GOARM=7 \
go build -ldflags "-X main.appVersion=$VER" -o $app
tar -acf $app.zip $app && rm -f $app
sha256sum $app.zip >> sha256-checksums)

app="youtubeuploader_linux_arm64"
echo "Generating $app ..."
(env GOOS=linux GOARCH=arm64 \
go build -ldflags "-X main.appVersion=$VER" -o $app
tar -acf $app.zip $app && rm -f $app
sha256sum $app.zip >> sha256-checksums)

app="youtubeuploader_linux_386"
echo "Generating $app ..."
(env GOOS=linux GOARCH=386 \
go build -ldflags "-X main.appVersion=$VER" -o $app
tar -acf $app.zip $app && rm -f $app
sha256sum $app.zip >> sha256-checksums)

app="youtubeuploader_linux_amd64"
echo "Generating $app ..."
(env GOOS=linux GOARCH=amd64 \
go build -ldflags "-X main.appVersion=$VER" -o $app
tar -acf $app.zip $app && rm -f $app
sha256sum $app.zip >> sha256-checksums)

app="youtubeuploader_windows_386"
echo "Generating $app ..."
(env GOOS=windows GOARCH=386 \
go build -ldflags "-X main.appVersion=$VER" -o $app.exe
tar -acf $app.zip $app.exe && rm -f $app.exe
sha256sum $app.zip >> sha256-checksums)

app="youtubeuploader_windows_amd64"
echo "Generating $app ..."
(env GOOS=windows GOARCH=amd64 \
go build -ldflags "-X main.appVersion=$VER" -o $app.exe
tar -acf $app.zip $app.exe && rm -f $app.exe
sha256sum $app.zip >> sha256-checksums)

app="youtubeuploader_mac_amd64"
echo "Generating $app ..."
(env GOOS=darwin GOARCH=amd64 \
go build -ldflags "-X main.appVersion=$VER" -o $app
tar -acf $app.zip $app && rm -f $app
sha256sum $app.zip >> sha256-checksums)

echo ":)"
