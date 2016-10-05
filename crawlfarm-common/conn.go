package crawlfarmCommon

import (
    "encoding/gob"
    "net"
    "time"
)


type Gob struct {
    Decoder *gob.Decoder
    Encoder *gob.Encoder
}

func RegisterGobs(){
    gob.Register(Packet{})
}

func GetGob(conn net.Conn) (Gob) {
    return Gob{ Encoder: gob.NewEncoder(conn), Decoder: gob.NewDecoder(conn) }
}

func handleError(err interface{}, disconnect chan bool){
    if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
        disconnect <- true
    } else if err != nil {
        disconnect <- true
    }
}

func sendPacket(conn net.Conn, enc *gob.Encoder, packet *Packet, disconnect chan bool) {
    conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
    if err := enc.Encode(packet); err != nil {
        handleError(err, disconnect)
    }
}

func receivePacket(conn net.Conn, dec *gob.Decoder, disconnect chan bool) (packet Packet, err error){
    conn.SetReadDeadline(time.Now().Add(time.Second * 2))
    if err := dec.Decode(&packet); err != nil {
        return packet, nil
    } else {
        handleError(err, disconnect)
        return packet, err
    }
}

func ReadSite(conn net.Conn, g Gob, disconnect chan bool) (site Site) {
    if packet, err := receivePacket(conn, g.Decoder, disconnect); err == nil && packet.Site != nil {
        site = *packet.Site
    }
    return
}

func ReadLoop(conn net.Conn, g Gob, links chan Link, results chan Result, urlCounts chan int, disconnect chan bool){
    for {
        if packet, err := receivePacket(conn, g.Decoder, disconnect); err == nil {
            if packet.Link != nil {
                links <- *packet.Link
            }

            if packet.Result != nil {
                results <- *packet.Result
            }

            if packet.Urls != nil {
                urlCounts <- *packet.Urls
            }
        } else if err != nil {
            disconnect <- true
        }
    }
}

func WriteSite(conn net.Conn, g Gob, site Site, disconnect chan bool){
    packet := &Packet { Site: &site }
    sendPacket(conn, g.Encoder, packet, nil)
}

func WriteLinks(conn net.Conn, g Gob, links chan Link, disconnect chan bool){
    for {
        select {
        case link := <-links:
            packet := &Packet{ Link: &link }
            sendPacket(conn, g.Encoder, packet, disconnect)
        }
    }
}

func WriteResults(conn net.Conn, g Gob, results chan Result, disconnect chan bool){
    for {
        select {
        case result := <-results:
            packet := &Packet{ Result: &result }
            sendPacket(conn, g.Encoder, packet, disconnect)
        }
    }
}

func WriteUrlCount(conn net.Conn, g Gob, counts chan int, disconnect chan bool){
    for {
        select {
        case count := <-counts:
            packet := &Packet{ Urls: &count }
            sendPacket(conn, g.Encoder, packet, disconnect)

        }
    }
}

func WriteFinished(conn net.Conn, g Gob, finished bool, disconnect chan bool){
    packet := &Packet{ Result: &Result{ JobFinished: finished }}
    sendPacket(conn, g.Encoder, packet, disconnect)
}
