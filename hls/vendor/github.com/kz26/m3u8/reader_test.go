/*
 Playlist parsing tests.

 Copyleft 2013-2014 Alexander I.Grafov aka Axel <grafov@gmail.com>

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU General Public License for more details.

 You should have received a copy of the GNU General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package m3u8

import (
	"bufio"
	"os"
	"testing"
)

func TestDecodeMasterPlaylist(t *testing.T) {
	f, err := os.Open("sample-playlists/master.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
		t.Fatal(err)
	}
	// check parsed values
	if p.ver != 3 {
		t.Errorf("Version of parsed playlist = %d (must = 3)", p.ver)
	}
	if len(p.Variants) != 5 {
		t.Error("Not all variants in master playlist parsed.")
	}
	// TODO check other values
	// fmt.Println(p.Encode().String())
}

func TestDecodeMasterPlaylistWithAlternatives(t *testing.T) {
	f, err := os.Open("sample-playlists/master-with-alternatives.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
		t.Fatal(err)
	}
	// check parsed values
	if p.ver != 3 {
		t.Errorf("Version of parsed playlist = %d (must = 3)", p.ver)
	}
	// if len(p.Variants) != 5 {
	// 	t.Fatal("Not all variants in master playlist parsed.")
	// }
	// TODO check other values
	//fmt.Println(p.Encode().String())
}

func TestDecodeMediaPlaylist(t *testing.T) {
	f, err := os.Open("sample-playlists/wowza-vod-chunklist.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewMediaPlaylist(5, 798)
	if err != nil {
		t.Fatalf("Create media playlist failed: %s", err)
	}
	err = p.DecodeFrom(bufio.NewReader(f), true)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("Playlist object: %+v\n", p)
	// check parsed values
	if p.ver != 3 {
		t.Errorf("Version of parsed playlist = %d (must = 3)", p.ver)
	}
	if p.TargetDuration != 12 {
		t.Errorf("TargetDuration of parsed playlist = %f (must = 12.0)", p.TargetDuration)
	}
	if !p.Closed {
		t.Error("This is a closed (VOD) playlist but Close field = false")
	}
	// TODO check other values…
	//fmt.Println(p.Encode().String()), stream.Name}
}

func TestDecodeMediaPlaylistWithWidevine(t *testing.T) {
	f, err := os.Open("sample-playlists/widevine-bitrate.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewMediaPlaylist(5, 798)
	if err != nil {
		t.Fatalf("Create media playlist failed: %s", err)
	}
	err = p.DecodeFrom(bufio.NewReader(f), true)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("Playlist object: %+v\n", p)
	// check parsed values
	if p.ver != 2 {
		t.Errorf("Version of parsed playlist = %d (must = 2)", p.ver)
	}
	if p.TargetDuration != 9 {
		t.Errorf("TargetDuration of parsed playlist = %f (must = 9.0)", p.TargetDuration)
	}
	// TODO check other values…
	//fmt.Printf("%+v\n", p.Key)
	//fmt.Println(p.Encode().String())
}

func TestDecodeMasterPlaylistWithAutodetection(t *testing.T) {
	print("test")
	f, err := os.Open("sample-playlists/master.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	m, listType, err := DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
		t.Fatal(err)
	}
	if listType != MASTER {
		t.Error("Sample not recognized as master playlist.")
	}
	mp := m.(*MasterPlaylist)
	// fmt.Printf(">%+v\n", mp)
	// for _, v := range mp.Variants {
	// 	fmt.Printf(">>%+v +v\n", v)
	// }
	//fmt.Println("Type below must be MasterPlaylist:")
	CheckType(t, mp)
}

func TestDecodeMediaPlaylistWithAutodetection(t *testing.T) {
	f, err := os.Open("sample-playlists/wowza-vod-chunklist.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	if err != nil {
		t.Fatal(err)
	}
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	if listType != MEDIA {
		t.Error("Sample not recognized as media playlist.")
	}
	// check parsed values
	if pp.TargetDuration != 12 {
		t.Errorf("TargetDuration of parsed playlist = %f (must = 12.0)", pp.TargetDuration)
	}

	if !pp.Closed {
		t.Error("This is a closed (VOD) playlist but Close field = false")
	}
	// TODO check other values…
	// fmt.Println(pp.Encode().String())
}
