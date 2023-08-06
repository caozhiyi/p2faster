package main

import (
	"os"
	"p2faster/peer"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/network"
)

var log = logging.Logger("ui")

type App struct {
	localId       string
	conn          *peer.BinaryConn
	trans         *peer.Transmission
	msgDispatcher *MsgDispatch
	recvFile      chan bool
	side          int

	app               fyne.App
	localIdLabel      *widget.Entry
	filePathEntry     *widget.Entry
	connectSteteLabel *widget.Label
	sendButton        *widget.Button
	cancelButton      *widget.Button
	recvButton        *widget.Button
	recvBox           *fyne.Container
	sendBox           *fyne.Container
}

func (a *App) Start() {
	logging.SetLogLevel("p2p-holepunch", "debug")
	logging.SetLogLevel("peer", "debug")
	logging.SetLogLevel("relay", "debug")
	logging.SetLogLevel("ui", "debug")
	a.recvFile = make(chan bool)
	a.side = SERVER

	a.conn = peer.CreateBinaryConn(
		a.onSendStream,
		a.onChatStream,
		a.onLocalId,
	)

	a.conn.Init()

	a.mainUI()
}

func (a *App) onChatStream(s network.Stream) {
	a.msgDispatcher = CreateMsgDispatch(s, a.side, a.onRecvFile, a.onSendFile)
	a.sendButton.Enable()
	a.recvButton.Enable()

	a.connectSteteLabel.SetText("connected")
	a.connectSteteLabel.Refresh()

	a.msgDispatcher.Start()
}

func (a *App) onSendStream(s network.Stream) {
	trans := peer.CreateTransmission(s)
	go trans.RecvFile(a.filePathEntry.Text)
}

func (a *App) onSendFile(recv bool) {
	if !recv {
		label := widget.NewLabel("peer cancel recv.")
		pop := widget.NewModalPopUp(label, test.Canvas())
		pop.Show()
		return
	}

	rw, err := a.conn.CreateSendStream()
	if err != nil {
		log.Errorf("create send file stream faied. err:%v", err)
		return
	}

	trans := peer.CreateTransmission(rw)
	trans.SendFile(a.filePathEntry.Text)
}

func (a *App) onRecvFile(name string, size int) bool {
	a.sendBox.Hide()
	a.recvBox.Show()
	a.cancelButton.Enable()
	a.recvButton.Enable()

	recv := <-a.recvFile

	a.sendBox.Show()
	a.recvBox.Hide()
	a.cancelButton.Disable()
	a.recvButton.Disable()

	// TODO check local file path
	return recv
}

func (a *App) onLocalId(id string) {
	a.localId = id
	if a.localIdLabel != nil {
		a.localIdLabel.SetText(a.localId)
		a.localIdLabel.Refresh()
	}
}

func (a *App) onConnButton(peerId string) {
	s, err := a.conn.Connect(peerId)
	if err != nil {
		a.connectSteteLabel.SetText("disconnected")
		a.connectSteteLabel.Refresh()
		return
	}
	a.side = CLIENT
	a.onChatStream(s)
}

func (a *App) mainUI() {
	a.app = app.New()
	w := a.app.NewWindow("p2faster")
	w.SetMaster()

	localIdTip := widget.NewLabel("local ID:")
	a.localIdLabel = widget.NewEntryWithData(binding.BindString(&a.localId))
	a.localIdLabel.Disable()
	localId := container.NewGridWithColumns(1, localIdTip, a.localIdLabel)

	peerIdLabel := widget.NewLabel("peer ID:")
	peerIdEntry := widget.NewEntry()
	peerId := container.NewGridWithColumns(1, peerIdLabel, peerIdEntry)

	connectButton := widget.NewButton("connect", func() {
		a.onConnButton(peerIdEntry.Text)
	})

	connectSteteLabel := widget.NewLabel("connect state:")
	a.connectSteteLabel = widget.NewLabel("disconnected")
	connectStete := container.NewGridWithColumns(2, connectSteteLabel, a.connectSteteLabel)
	connection := container.NewGridWithColumns(2, connectButton, connectStete)

	filePathLabel := widget.NewLabel("file path:")
	a.filePathEntry = widget.NewEntry()
	filePath := container.NewGridWithColumns(1, filePathLabel, a.filePathEntry)
	a.cancelButton = widget.NewButton("cancel", func() {
		log.Infof("cancel recv file. path:%s", a.filePathEntry.Text)
		a.recvFile <- false
	})
	a.cancelButton.Disable()
	a.recvButton = widget.NewButton("receive", func() {
		log.Infof("start recv file. path:%s", a.filePathEntry.Text)
		a.recvFile <- true
	})
	a.recvButton.Disable()
	a.recvBox = container.NewVBox(a.cancelButton, a.recvButton)
	a.recvBox.Hide()

	a.sendButton = widget.NewButton("send", func() {
		log.Infof("start send file. path:%s", a.filePathEntry.Text)
		fileInfo, err := os.Stat(a.filePathEntry.Text)
		if err != nil {
			label := widget.NewLabel("open file failed.")
			pop := widget.NewModalPopUp(label, test.Canvas())
			pop.Show()
			return
		}
		a.msgDispatcher.ConferSendFile(fileInfo.Name(), int(fileInfo.Size()))
	})
	a.sendButton.Disable()
	a.sendBox = container.NewVBox(a.sendButton)

	sendGrid := container.NewGridWithColumns(2, filePath, a.sendBox, a.recvBox)

	w.SetContent(container.NewVBox(
		localId,
		peerId,
		connection,
		sendGrid))
	w.Resize(fyne.NewSize(460, 360))
	w.FixedSize()

	w.ShowAndRun()
}
