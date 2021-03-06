package tobubus

import (
	"github.com/shibukawa/mockconn"
	"testing"
	"time"
)

func TestPluginConnect(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	sessionID := plugin.sessions.getUniqueSessionID() + 1
	socket.SetExpectedActions(
		mockconn.Write(archiveMessage(ConnectClient, sessionID, []byte("github.com/shibukawa/tobubus/1"))),
		mockconn.Read(archiveMessage(ResultOK, sessionID, nil)),
	)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
	}()
	plugin.connect()
	socket.Verify()
}

func TestPluginConnectError(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	sessionID := plugin.sessions.getUniqueSessionID() + 1
	socket.SetExpectedActions(
		mockconn.Write(archiveMessage(ConnectClient, sessionID, []byte("github.com/shibukawa/tobubus/1"))),
		mockconn.Read(archiveMessage(ResultNG, sessionID, nil)),
		mockconn.Close(),
	)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
	}()
	plugin.connect()
	socket.Verify()
}

func TestPluginUnregister(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	sessionID := plugin.sessions.getUniqueSessionID() + 1
	socket.SetExpectedActions(
		mockconn.Write(archiveMessage(CloseClient, sessionID, nil)),
		mockconn.Read(archiveMessage(ResultOK, sessionID, nil)),
	)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
	}()
	plugin.Close()
	socket.Verify()
}

func TestPluginConfirmPath(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	sessionID := plugin.sessions.getUniqueSessionID() + 1
	socket.SetExpectedActions(
		mockconn.Write(archiveMessage(ConfirmPath, sessionID, []byte("/image/reader"))),
		mockconn.Read(archiveMessage(ResultOK, sessionID, nil)),
	)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
	}()
	if !plugin.ConfirmPath("/image/reader") {
		t.Error("result should be true")
	}
	socket.Verify()
}

func TestPluginPublishThenConnect(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	sessionID := plugin.sessions.getUniqueSessionID() + 1
	socket.SetExpectedActions(
		mockconn.Write(archiveMessage(ConnectClient, sessionID, []byte("github.com/shibukawa/tobubus/1"))),
		mockconn.Read(archiveMessage(ResultOK, sessionID, nil)),
		mockconn.Write(archiveMessage(Publish, sessionID+1, []byte("/image/reader"))),
		mockconn.Read(archiveMessage(ResultOK, sessionID+1, nil)),
	)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
	}()
	obj := testStruct{result: "ok"}
	err := plugin.Publish("/image/reader", &obj)
	if err != nil {
		t.Errorf("error should be nil, but %v", err)
	}
	err = plugin.connect()
	if err != nil {
		t.Errorf("error should be nil, but %v", err)
	}
	socket.Verify()
}

func TestPluginConfirmPathNG(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	sessionID := plugin.sessions.getUniqueSessionID() + 1
	socket.SetExpectedActions(
		mockconn.Write(archiveMessage(ConfirmPath, sessionID, []byte("/image/reader"))),
		mockconn.Read(archiveMessage(ResultNG, sessionID, nil)),
	)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
	}()
	if plugin.ConfirmPath("/image/reader") {
		t.Error("result should be false")
	}
	socket.Verify()
}

func TestPluginCallMethod(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	sessionID := plugin.sessions.getUniqueSessionID() + 1
	send, _ := archiveMethodCallMessage(CallMethod, sessionID, "/image/reader", "open", []interface{}{"image.png"})
	receive, _ := archiveMethodCallMessage(ReturnMethod, sessionID, "", "", []interface{}{"ok"})
	socket.SetExpectedActions(
		mockconn.Write(send),
		mockconn.Read(receive),
	)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
	}()
	result, err := plugin.Call("/image/reader", "open", "image.png")
	if err != nil {
		t.Errorf("err should be nil, %v", err)
	} else if len(result) != 1 {
		t.Errorf("result count should be 1, but %d", len(result))
	} else if result[0].(string) != "ok" {
		t.Errorf("result error: %v", result[0])
	}
	socket.Verify()
}

func TestPluginCallLocalMethod(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	sessionID := plugin.sessions.getUniqueSessionID() + 1
	socket.SetExpectedActions(
		mockconn.Write(archiveMessage(ConnectClient, sessionID, []byte("github.com/shibukawa/tobubus/1"))),
		mockconn.Read(archiveMessage(ResultOK, sessionID, nil)),
		mockconn.Write(archiveMessage(Publish, sessionID+1, []byte("/image/reader"))),
		mockconn.Read(archiveMessage(ResultOK, sessionID+1, nil)),
	)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
	}()

	obj := testStruct{result: "ok"}
	err := plugin.Publish("/image/reader", &obj)
	if err != nil {
		t.Error("result should not be nil")
	}
	err = plugin.connect()
	if err != nil {
		t.Errorf("error should be nil, but %v", err)
	}
	result, err := plugin.Call("/image/reader", "TestMethod", "test value")
	if err != nil {
		t.Errorf("error should be nil, but %v", err)
	}
	if len(result) != 1 {
		t.Errorf("obj.TestMethod should return one value, but %d result is returned", len(result))
	} else if result[0] != "ok" {
		t.Errorf("obj.TestMethod should return 'ok' but '%v' is returnd", result[0])
	}
	if len(obj.args) != 1 {
		t.Errorf("obj.TestMethod should be called with one argument, but %d argument is passed", len(obj.args))
	} else if obj.args[0] != "test value" {
		t.Errorf("obj.args[0] should be 'image.png', but %v", obj.args[0])
	}

	socket.Verify()
}

func TestPluginMethodCalledFromHost(t *testing.T) {
	plugin, socket := newPluginForTest("pipe.test", "github.com/shibukawa/tobubus/1", t)
	hostSessionID := uint32(45)
	receive, _ := archiveMethodCallMessage(CallMethod, hostSessionID, "/image/reader", "TestMethod", []interface{}{"image.png"})
	send, _ := archiveMethodCallMessage(ReturnMethod, hostSessionID, "", "", []interface{}{"ok"})

	sessionID := plugin.sessions.getUniqueSessionID() + 1

	socket.SetExpectedActions(
		mockconn.Write(archiveMessage(ConnectClient, sessionID, []byte("github.com/shibukawa/tobubus/1"))),
		mockconn.Read(archiveMessage(ResultOK, sessionID, nil)),
		mockconn.Write(archiveMessage(Publish, sessionID+1, []byte("/image/reader"))),
		mockconn.Read(archiveMessage(ResultOK, sessionID+1, nil)),
		mockconn.Read(receive),
		mockconn.Write(send),
	)
	wait := make(chan string)
	go func() {
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
		time.Sleep(time.Millisecond)
		plugin.receiveMessage()
		time.Sleep(time.Millisecond)
		wait <- "done"
	}()
	obj := testStruct{result: "ok"}
	err := plugin.Publish("/image/reader", &obj)
	if err != nil {
		t.Error("result should not be nil")
	}
	err = plugin.connect()
	if err != nil {
		t.Errorf("error should be nil, but %v", err)
	}
	// Receive method call from host
	<-wait

	if len(obj.args) != 1 {
		t.Errorf("obj.TestMethod should be called with one argument, but %d argument is passed", len(obj.args))
	} else if obj.args[0] != "image.png" {
		t.Errorf("obj.args[0] should be 'image.png', but %v", obj.args[0])
	}
	socket.Verify()
}
