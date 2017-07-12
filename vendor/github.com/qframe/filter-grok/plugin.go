package qfilter_grok

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"github.com/vjeantet/grok"
	"github.com/zpatrick/go-config"

	"github.com/qnib/qframe-types"
	"github.com/qnib/qframe-utils"
)

const (
	version = "0.1.12"
	pluginTyp = "filter"
	pluginPkg = "grok"
	defPatternDir = "/etc/grok-patterns"
)

type Plugin struct {
	qtypes.Plugin
	grok    *grok.Grok
	pattern string
}

func (p *Plugin) GetOverwriteKeys() []string {
	inStr, err := p.CfgString("overwrite-keys")
	if err != nil {
		inStr = ""
	}
	return strings.Split(inStr, ",")
}

func New(qChan qtypes.QChan, cfg *config.Config, name string) (p Plugin, err error) {
	p = Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg,  name, version),
	}
	return p, err
}

func (p *Plugin) Match(str string) (map[string]string, bool) {
	match := true
	p.Log("trace", fmt.Sprintf("Match '%s' against '%s'", str, p.pattern))
	val, _ := p.grok.Parse(p.pattern, str)
	keys := reflect.ValueOf(val).MapKeys()
	if len(keys) == 0 {
		match = false
	}
	return val, match
}

func (p *Plugin) GetPattern() string {
	return p.pattern
}

func (p *Plugin) InitGrok() {
	p.grok, _ = grok.New()
	var err error
	p.pattern, err = p.CfgString("pattern")
	if err != nil {
		p.Log("fatal", "Could not find pattern in config")
	}
	pDir, err := p.CfgString("pattern-dir")
	if err != nil {
		if _, err := os.Stat(defPatternDir); err == nil {
			pDir = defPatternDir
			p.Log("info", fmt.Sprintf("Add patterns from DEFAULT directory '%s'", pDir))
		}
	} else {
		p.Log("info", fmt.Sprintf("Add patterns from directory '%s'", pDir))
	}
	if _, err := os.Stat(pDir); err != nil {
		p.Log("error", fmt.Sprintf("Patterns directory does not exist '%s'", pDir))
	} else {
		p.grok.AddPatternsFromPath(pDir)
	}
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start grok filter v%s", p.Version))
	p.InitGrok()
	p.MyID = qutils.GetGID()
	bg := p.QChan.Data.Join()
	msgKey := p.CfgStringOr("overwrite-message-key", "")
	for {
		val := bg.Recv()
		switch val.(type) {
		case qtypes.Message:
			qm := val.(qtypes.Message)
			if p.StopProcessingMessage(qm, false) {
				continue
			}
			qm.AppendSource(p.Name)
			var kv map[string]string
			kv, qm.SourceSuccess = p.Match(qm.Message)
			if qm.SourceSuccess {
				p.Log("debug", fmt.Sprintf("Matched pattern '%s'", p.pattern))
				for k,v := range kv {
					p.Log("debug", fmt.Sprintf("    %15s: %s", k,v ))
					qm.KV[k] = v
					if msgKey == k {
						qm.Message = v
					}
				}
			}/* else {
				p.Log("debug", fmt.Sprintf("No match of '%s' for message '%s'", p.pattern, qm.Message))
			}*/
			p.QChan.Data.Send(qm)
		}
	}
}
