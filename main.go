package main

import (
    "errors"
    "io"
    "os"
    "fmt"
    "io/ioutil"
    "bytes"
    "encoding/json"

    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/crypto/ssh/terminal"
    // "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/rlp"
    "github.com/ethereum/go-ethereum/common/hexutil"
)

var (
    // IstanbulDigest represents a hash of "Istanbul practical byzantine fault tolerance"
    // to identify whether the block is from Istanbul consensus engine
    IstanbulDigest = common.HexToHash("0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365")

    IstanbulExtraVanity = 32 // Fixed number of extra-data bytes reserved for validator vanity
    IstanbulExtraSeal   = 65 // Fixed number of extra-data bytes reserved for validator seal

    // ErrInvalidIstanbulHeaderExtra is returned if the length of extra-data is less than 32 bytes
    ErrInvalidIstanbulHeaderExtra = errors.New("invalid istanbul header extra-data")
)

type IstanbulExtra struct {
    Validators    []common.Address
    Seal          []byte
    CommittedSeal [][]byte
}


// EncodeRLP serializes ist into the Ethereum RLP format.
func (ist *IstanbulExtra) EncodeRLP(w io.Writer) error {
    return rlp.Encode(w, []interface{}{
        ist.Validators,
        ist.Seal,
        ist.CommittedSeal,
    })
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (ist *IstanbulExtra) DecodeRLP(s *rlp.Stream) error {
    var istanbulExtra struct {
        Validators    []common.Address
        Seal          []byte
        CommittedSeal [][]byte
    }
    if err := s.Decode(&istanbulExtra); err != nil {
        return err
    }
    ist.Validators, ist.Seal, ist.CommittedSeal = istanbulExtra.Validators, istanbulExtra.Seal, istanbulExtra.CommittedSeal
    return nil
}

// ExtractIstanbulExtra extracts all values of the IstanbulExtra from the header. It returns an
// error if the length of the given extra-data is less than 32 bytes or the extra-data can not
// be decoded.
func ExtractIstanbulExtra(h *types.Header) (*IstanbulExtra, error) {
    if len(h.Extra) < IstanbulExtraVanity {
        return nil, ErrInvalidIstanbulHeaderExtra
    }

    var istanbulExtra *IstanbulExtra
    err := rlp.DecodeBytes(h.Extra[IstanbulExtraVanity:], &istanbulExtra)
    if err != nil {
        return nil, err
    }
    return istanbulExtra, nil
}

// IstanbulFilteredHeader returns a filtered header which some information (like seal, committed seals)
// are clean to fulfill the Istanbul hash rules. It returns nil if the extra-data cannot be
// decoded/encoded by rlp.
func IstanbulFilteredHeader(h *types.Header, keepSeal bool) *types.Header {
    newHeader := types.CopyHeader(h)
    istanbulExtra, err := ExtractIstanbulExtra(newHeader)
    if err != nil {
        return nil
    }

    if !keepSeal {
        istanbulExtra.Seal = []byte{}
    }
    istanbulExtra.CommittedSeal = [][]byte{}

    payload, err := rlp.EncodeToBytes(&istanbulExtra)
    if err != nil {
        return nil
    }

    newHeader.Extra = append(newHeader.Extra[:IstanbulExtraVanity], payload...)

    return newHeader
}


func Encode(vanity string, validators []common.Address) (string, error) {
    newVanity, err := hexutil.Decode(vanity)
    if err != nil {
        return "", err
    }

    if len(newVanity) < IstanbulExtraVanity {
        newVanity = append(newVanity, bytes.Repeat([]byte{0x00}, IstanbulExtraVanity-len(newVanity))...)
    }
    newVanity = newVanity[:IstanbulExtraVanity]

    ist := &IstanbulExtra{
        Validators:    validators,
        Seal:          make([]byte, IstanbulExtraSeal),
        CommittedSeal: [][]byte{},
    }

    payload, err := rlp.EncodeToBytes(&ist)
    if err != nil {
        return "", err
    }

    return "0x" + common.Bytes2Hex(append(newVanity, payload...)), nil
}

func ReadSTDINtoAddress(data []byte) ([]common.Address, error) {
    var addrs []string
    var addrsDec []common.Address
    json.Unmarshal(data, &addrs)
    for i := 0; i < len(addrs); i++ {
        addrsDec = append(addrsDec, common.HexToAddress(addrs[i]))
    }
    return addrsDec, nil
}

func main() {
    if ! terminal.IsTerminal(0) {
        fmt.Print("Extradata for addrs: ")
        b, _ := ioutil.ReadAll(os.Stdin)
        addrs, _ := ReadSTDINtoAddress(b)
        extraData, _ := Encode("0x00", addrs)
        fmt.Printf("{%s}\n", extraData)
    } else {
        fmt.Println("no piped data\n")
    }
}