#!/usr/bin/env python3
"""Build a deterministic, offline-only SCTP module candidate from a clean tree.

Reconstruct the input from the pinned base and ordered evidence patches first.
No tag, VCS commit, remote upload, replace directive or public version is created.
"""
import argparse
import hashlib
import json
from pathlib import Path
import zipfile

MODULE = "github.com/skill215/sctp"
VERSION = "v0.0.0-diameterkit.u5.62"


CONSUMER = r'''package consumer
import (
 "context"
 "encoding/binary"
 "errors"
 "net"
 "testing"
 "time"
 "github.com/skill215/sctp"
)
func TestInstalledMessageRoundTrip(t *testing.T) {
 ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second);defer cancel()
 m:=sctp.MessageOptions{ReceiveInfo:sctp.ReceiveInfoSNDRCV,NoDelay:true,AllowFragmentation:true}
 cfg:=sctp.SocketConfig{Messages:&m}
 ln,e:=cfg.Listen("sctp4",&sctp.SCTPAddr{IPAddrs:[]net.IPAddr{{IP:net.IPv4(127,0,0,1)}}});if e!=nil{t.Fatal(e)};defer ln.Close()
 addr,e:=sctp.ResolveSCTPAddrContext(ctx,"sctp4",ln.Addr().String());if e!=nil{t.Fatal(e)}
 c,e:=cfg.DialContext(ctx,"sctp4",nil,addr);if e!=nil{t.Fatal(e)};defer c.Close()
 s,e:=ln.AcceptSCTP();if e!=nil{t.Fatal(e)};defer s.Close()
 state,e:=s.GetMessageOptions();if e!=nil||state.Options!=m{t.Fatal(state,e)}
 limits,e:=sctp.MessageIORequirements(m);if e!=nil{t.Fatal(e)}
 payload:=[]byte("installed-candidate")
 if n,e:=c.WriteMsgContext(ctx,payload,sctp.SndInfo{PPID:46});e!=nil||n!=len(payload){t.Fatal(n,e)}
 d:=make([]byte,64);b:=make([]byte,limits.NotificationMin);o:=make([]byte,limits.OOBMin)
 r,e:=s.ReadEvent(ctx,d,b,o);if e!=nil||r.Kind!=sctp.ReadData||!r.InfoPresent||r.Info.PPID!=46||string(d[:r.N])!=string(payload){t.Fatal(r,e)}
 h:=make([]byte,12);binary.NativeEndian.PutUint16(h,uint16(sctp.SCTP_SHUTDOWN_EVENT));binary.NativeEndian.PutUint32(h[4:],12)
 event,e:=sctp.ParseNotification(h);if e!=nil||event.Kind!=sctp.NotificationShutdown{t.Fatal(event,e)}
 cancel();if _,e:=s.ReadEvent(ctx,d,b,o);!errors.Is(e,context.Canceled){t.Fatal(e)}
}
'''


def main():
    parser = argparse.ArgumentParser(__doc__)
    parser.add_argument("source", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    source, output = args.source.resolve(), args.output.resolve()
    if source == output or source in output.parents:
        parser.error("output must be outside source")
    files = []
    for path in sorted(source.rglob("*")):
        if path.is_symlink():
            parser.error(f"symlink not accepted: {path}")
        if not path.is_file():
            continue
        relative = path.relative_to(source).as_posix()
        if ".git" in path.relative_to(source).parts:
            continue
        files.append((relative, path.read_bytes()))
    contents = dict(files)
    if not contents.get("go.mod", b"").startswith(f"module {MODULE}\n".encode()):
        parser.error("unexpected module")
    proxy = output / "proxy" / MODULE / "@v"
    proxy.mkdir(parents=True, exist_ok=True)
    archive = proxy / (VERSION + ".zip")
    with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as z:
        for name, data in files:
            info = zipfile.ZipInfo(f"{MODULE}@{VERSION}/{name}", (1980, 1, 1, 0, 0, 0))
            info.create_system = 3
            info.external_attr = 0o100644 << 16
            info.compress_type = zipfile.ZIP_DEFLATED
            z.writestr(info, data)
    (proxy / (VERSION + ".mod")).write_bytes(contents["go.mod"])
    (proxy / (VERSION + ".info")).write_text(json.dumps({"Version": VERSION, "Time": "2026-11-09T00:00:00Z"}) + "\n")
    (proxy / "list").write_text(VERSION + "\n")
    consumer = output / "consumer"
    consumer.mkdir(exist_ok=True)
    (consumer / "go.mod").write_text(f"module example.com/sctp-offline-consumer\n\ngo 1.26.0\nrequire {MODULE} {VERSION}\n")
    (consumer / "candidate_test.go").write_text(CONSUMER)
    manifest = {
        "module": MODULE, "offline_version": VERSION,
        "publication": "none; version is a local file-proxy test label",
        "zip_sha256": hashlib.sha256(archive.read_bytes()).hexdigest(),
        "files": [{"path": name, "sha256": hashlib.sha256(data).hexdigest()} for name, data in files],
    }
    (output / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(json.dumps({"zip": str(archive), "files": len(files), "sha256": manifest["zip_sha256"]}))


if __name__ == "__main__":
    main()
