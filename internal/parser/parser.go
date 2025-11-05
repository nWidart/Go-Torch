package parser

import (
	"regexp"
	"strings"
	"time"

	"GoTorch/internal/types"
)

// Parser compiles and applies regex patterns to log lines to emit normalized Events.
// Patterns are based on initial Python reference and refined with real logs.

type Parser struct {
	bagInit    *regexp.Regexp
	bagMod     *regexp.Regexp
	transition *regexp.Regexp // captures NextSceneName path
	tsPrefix   *regexp.Regexp // captures timestamp components
}

func New() *Parser {
	// Flexible prefix: allow optional [..][..] at start followed by GameLog: Display: [Game]
	prefix := `(?:\[.*?\]){1,3}\s*GameLog: Display: \[Game\]\s*`

	bagInit := regexp.MustCompile(prefix + `BagMgr@:InitBagData\s+PageId = (\d+)\s+SlotId = (\d+)\s+ConfigBaseId = (\d+)\s+Num = (\d+)`)
	bagMod := regexp.MustCompile(prefix + `BagMgr@:Modfy BagItem\s+PageId = (\d+)\s+SlotId = (\d+)\s+ConfigBaseId = (\d+)\s+Num = (\d+)`)

	// Transition with NextSceneName = World'/Game/Art/Maps...'
	transition := regexp.MustCompile(`PageApplyBase@ _UpdateGameEnd: .*?NextSceneName = World'(/Game/Art/Maps[^']*)'`)

	// Timestamp prefix: [YYYY.MM.DD-HH.MM.SS:ms][...]
	tsPrefix := regexp.MustCompile(`^\[(\d{4})\.(\d{2})\.(\d{2})-(\d{2})\.(\d{2})\.(\d{2}):(\d{3})\]`)

	return &Parser{bagInit: bagInit, bagMod: bagMod, transition: transition, tsPrefix: tsPrefix}
}

const refugePath = "/Game/Art/Maps/01SD/XZ_YuJinZhiXiBiNanSuo200/XZ_YuJinZhiXiBiNanSuo200.XZ_YuJinZhiXiBiNanSuo200"

// Parse attempts to parse a line into an Event. Returns nil if unrecognized.
func (p *Parser) Parse(line string) *types.Event {
	line = strings.TrimRight(line, "\r\n")
	ts := p.parseTimestamp(line)
	if m := p.bagInit.FindStringSubmatch(line); m != nil {
		return &types.Event{Kind: types.EventBagInit, Time: ts, Line: line, Bag: parseBag(m)}
	}
	if m := p.bagMod.FindStringSubmatch(line); m != nil {
		return &types.Event{Kind: types.EventBagMod, Time: ts, Line: line, Bag: parseBag(m)}
	}
	if m := p.transition.FindStringSubmatch(line); m != nil {
		path := m[1]
		if strings.HasPrefix(path, "/Game/Art/Maps/") {
			if path == refugePath {
				return &types.Event{Kind: types.EventMapEnd, Time: ts, Line: line}
			}
			return &types.Event{Kind: types.EventMapStart, Time: ts, Line: line}
		}
	}
	return nil
}

func (p *Parser) parseTimestamp(line string) time.Time {
	m := p.tsPrefix.FindStringSubmatch(line)
	if m == nil {
		return time.Now()
	}
	y := atoi(m[1])
	mon := atoi(m[2])
	d := atoi(m[3])
	h := atoi(m[4])
	min := atoi(m[5])
	s := atoi(m[6])
	ms := atoi(m[7])
	return time.Date(y, time.Month(mon), d, h, min, s, ms*1e6, time.Local)
}

func parseBag(matches []string) *types.BagEvent {
	// matches[0] is the full match
	page := atoi(matches[1])
	slot := atoi(matches[2])
	cfg := atoi(matches[3])
	num := atoi(matches[4])
	return &types.BagEvent{PageID: page, SlotID: slot, ConfigBaseID: cfg, Num: num}
}

func atoi(s string) int {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			continue
		}
		n = n*10 + int(c-'0')
	}
	return n
}
