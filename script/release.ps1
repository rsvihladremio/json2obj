
# script/release: build binaries in all supported platforms and upload them with the gh client

param(
     $VERSION
)

Set-Location "$PSScriptRoot\.."


# this is also set in script/build and is a copy paste
$GIT_SHA=@(git rev-parse --short HEAD)
$LDFLAGS="-X github.com/rsvihladremio/json2obj/cmd.GitSha=$GIT_SHA -X github.com/rsvihladremio/json2obj/cmd.Version=$VERSION"

Write-Output "Cleaning bin folder"
Get-Date
.\script\clean

Write-Output "Building linux-amd64"
Get-Date
$Env:GOOS='linux' 
$Env:GOARCH='amd64' 
go build -ldflags "$LDFLAGS" -o ./bin/json2obj
zip .\bin\json2obj-$VERSION-linux-amd64.zip .\bin\json2obj
Write-Output "Building linux-arm64"
Get-Date
$Env:GOOS='linux' 
$Env:GOARCH='arm64'
go build -ldflags "$LDFLAGS" -o ./bin/json2obj
zip .\bin\json2obj-$VERSION-linux-arm64.zip .\bin\json2obj
Write-Output "Building darwin-os-x-amd64"
Get-Date
$Env:GOOS='darwin' 
$Env:GOARCH='amd64'
go build -ldflags "$LDFLAGS" -o ./bin/json2obj
zip .\bin\json2obj-$VERSION-darwin-amd64.zip .\bin\json2obj
Write-Output "Building darwin-os-x-arm64"
Get-Date
$Env:GOOS='darwin' 
$Env:GOARCH='arm64'
go build -ldflags "$LDFLAGS" -o ./bin/json2obj
zip .\bin\json2obj-$VERSION-darwin-arm64.zip .\bin\json2obj
Write-Output "Building windows-amd64"
Get-Date
$Env:GOOS='windows' 
$Env:GOARCH='amd64'
go build -ldflags "$LDFLAGS" -o ./bin/json2obj.exe
zip .\bin\json2obj-$VERSION-windows-amd64.zip .\bin\json2obj.exe
Write-Output "Building windows-arm64"
Get-Date
$Env:GOOS='windows' 
$Env:GOARCH='arm64'
go build -ldflags "$LDFLAGS" -o ./bin/json2obj.exe
zip .\bin\json2obj-$VERSION-windows-arm64.zip .\bin\json2obj.exe

Remove-Item -Path Env:\GOOS
Remove-Item -Path Env:\GOARCH 
gh release create $VERSION --title $VERSION -d -F changelog.md .\bin\json2obj-$VERSION-windows-arm64.zip .\bin\json2obj-$VERSION-windows-amd64.zip .\bin\json2obj-$VERSION-darwin-arm64.zip .\bin\json2obj-$VERSION-darwin-amd64.zip .\bin\json2obj-$VERSION-linux-arm64.zip .\bin\json2obj-$VERSION-linux-amd64.zip 
 