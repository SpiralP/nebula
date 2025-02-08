package nebula

import (
	"fmt"
	"net/netip"
	"sort"
	"strings"
)

type edge struct {
	from    string
	to      string
	dual    bool
	invalid bool
}

func RenderControlsHostmaps(controls ...*Control) string {
	interfaces := make([]*Interface, len(controls))
	for i, c := range controls {
		interfaces[i] = c.f
	}
	return RenderHostmaps(false, interfaces...)
}

func RenderHostmaps(mermaid bool, interfaces ...*Interface) string {
	var lines []*edge
	r := ""
	if mermaid {
		r += "graph TB\n"
	} else {
		r += "digraph G {\n"
		r += "\tranksep=1\n"
		r += "\tnode [shape=box]\n"
	}

	for _, c := range interfaces {
		sr, se := renderHostmap(mermaid, c)
		r += sr
		for _, e := range se {
			add := true

			// Collapse duplicate edges into a bi-directionally connected edge
			for _, ge := range lines {
				if e.to == ge.from && e.from == ge.to {
					add = false
					ge.dual = true
					break
				}
			}

			if add {
				lines = append(lines, e)
			}
		}
	}

	sort.SliceStable(lines, func(i, j int) bool {
		if lines[i].from == lines[j].from {
			return lines[i].to < lines[j].to
		}
		return lines[i].from < lines[j].from
	})
	for _, line := range lines {
		if mermaid {
			if line.dual {
				r += fmt.Sprintf("\t%v <--> %v\n", line.from, line.to)
			} else {
				r += fmt.Sprintf("\t%v --> %v\n", line.from, line.to)
			}
		} else {
			if line.dual {
				if line.invalid {
					r += fmt.Sprintf("\t%v -> %v [dir=both color=red]\n", line.from, line.to)
				} else {
					r += fmt.Sprintf("\t%v -> %v [dir=both]\n", line.from, line.to)
				}
			} else {
				if line.invalid {
					r += fmt.Sprintf("\t%v -> %v [color=red]\n", line.from, line.to)
				} else {
					r += fmt.Sprintf("\t%v -> %v\n", line.from, line.to)
				}
			}
		}
	}

	if !mermaid {
		r += "}\n"
	}

	return r
}
func renderHostmap(mermaid bool, f *Interface) (string, []*edge) {
	var lines []string
	var globalLines []*edge

	clusterName := strings.Trim(f.pki.GetCertState().Certificate.Details.Name, " ")
	clusterVpnIp := f.pki.GetCertState().Certificate.Details.Ips[0].IP
	r := ""
	if mermaid {
		r += fmt.Sprintf("\tsubgraph %s[\"%s (%s)\"]\n", clusterName, clusterName, clusterVpnIp)
	} else {
		r += fmt.Sprintf("\tsubgraph cluster_%s {\n", clusterName)
		r += fmt.Sprintf("\t\tlabel=\"%s (%s)\"\n", clusterName, clusterVpnIp)
	}

	f.hostMap.RLock()
	defer f.hostMap.RUnlock()

	// Draw the vpn to index nodes
	if mermaid {
		r += fmt.Sprintf("\t\tsubgraph %s.hosts[\"Hosts\"]\n", clusterName)
	} else {
		r += fmt.Sprintf("\t\tsubgraph cluster_%s_hosts {\n", clusterName)
		r += "\t\t\tlabel=\"Hosts\"\n"
	}
	hosts := sortedHosts(f.hostMap.Hosts)
	for _, vpnIp := range hosts {
		hi := f.hostMap.Hosts[vpnIp]
		if mermaid {
			r += fmt.Sprintf("\t\t\t%v.%v[\"%v\"]\n", clusterName, vpnIp, vpnIp)
			lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, vpnIp, clusterName, hi.localIndexId))
		} else {
			r += fmt.Sprintf("\t\t\t\"%v_%v\" [label=\"%v\"]\n", clusterName, vpnIp, vpnIp)
			if !hi.remote.IsValid() {
				lines = append(lines, fmt.Sprintf("\"%v_%v\" -> \"%v_%v\" [color=red]", clusterName, vpnIp, clusterName, hi.localIndexId))
			} else {
				lines = append(lines, fmt.Sprintf("\"%v_%v\" -> \"%v_%v\"", clusterName, vpnIp, clusterName, hi.localIndexId))
			}
		}

		for _, relayIp := range hi.relayState.CopyRelayIps() {
			if mermaid {
				lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, vpnIp, clusterName, relayIp))
			} else {
				lines = append(lines, fmt.Sprintf("\"%v_%v\" -> \"%v_%v\"", clusterName, vpnIp, clusterName, relayIp))
			}
		}

		for _, relayIp := range hi.relayState.CopyRelayForIdxs() {
			if mermaid {
				lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, vpnIp, clusterName, relayIp))
			} else {
				lines = append(lines, fmt.Sprintf("\"%v_%v\" -> \"%v_%v\"", clusterName, vpnIp, clusterName, relayIp))
			}
		}
	}
	if mermaid {
		r += "\t\tend\n"
	} else {
		r += "\t\t}\n"
	}

	// Draw the relay hostinfos
	if len(f.hostMap.Relays) > 0 {
		if mermaid {
			r += fmt.Sprintf("\t\tsubgraph %s.relays[\"Relays\"]\n", clusterName)
		} else {
			r += fmt.Sprintf("\t\tsubgraph cluster_%s_relays {\n", clusterName)
			r += "\t\t\tlabel=\"Relays\"\n"
		}
		indexes := sortedIndexes(f.hostMap.Relays)
		for _, relayIndex := range indexes {
			hi := f.hostMap.Relays[relayIndex]
			if mermaid {
				r += fmt.Sprintf("\t\t\t%v.%v[\"%v\\n%v\"]\n", clusterName, relayIndex, hi.vpnIp, relayIndex)
				lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, relayIndex, clusterName, hi.localIndexId))
			} else {
				r += fmt.Sprintf("\t\t\t\"%v_%v\" [label=\"%v\\n%v\"]\n", clusterName, relayIndex, hi.vpnIp, relayIndex)
				lines = append(lines, fmt.Sprintf("\"%v_%v\" -> \"%v_%v\"", clusterName, relayIndex, clusterName, hi.localIndexId))
			}
		}
		if mermaid {
			r += "\t\tend\n"
		} else {
			r += "\t\t}\n"
		}
	}

	// Draw the local index to relay or remote index nodes
	if mermaid {
		r += fmt.Sprintf("\t\tsubgraph indexes.%s[\"Indexes\"]\n", clusterName)
	} else {
		r += fmt.Sprintf("\t\tsubgraph cluster_%s_indexes {\n", clusterName)
		r += "\t\t\tlabel=\"Indexes\"\n"
	}
	indexes := sortedIndexes(f.hostMap.Indexes)
	for _, idx := range indexes {
		hi, ok := f.hostMap.Indexes[idx]
		if ok {
			remoteClusterName := strings.Trim(hi.GetCert().Details.Name, " ")
			if mermaid {
				r += fmt.Sprintf("\t\t\t%v.%v[\"%v\\n%v\"]\n", clusterName, idx, hi.remote, idx)
				globalLines = append(globalLines, &edge{
					from: fmt.Sprintf("%v.%v", clusterName, idx),
					to:   fmt.Sprintf("%v.%v", remoteClusterName, hi.remoteIndexId),
				})
			} else {
				if !hi.remote.IsValid() {
					r += fmt.Sprintf("\t\t\t\"%v_%v\" [label=\"%v\\n%v\" color=red]\n", clusterName, idx, hi.remote, idx)
				} else {
					r += fmt.Sprintf("\t\t\t\"%v_%v\" [label=\"%v\\n%v\"]\n", clusterName, idx, hi.remote, idx)
				}
				globalLines = append(globalLines, &edge{
					from:    fmt.Sprintf("\"%v_%v\"", clusterName, idx),
					to:      fmt.Sprintf("\"%v_%v\"", remoteClusterName, hi.remoteIndexId),
					invalid: !hi.remote.IsValid(),
				})
			}
		}
	}
	if mermaid {
		r += "\t\tend\n"
	} else {
		r += "\t\t}\n"
	}

	// Add the edges inside this host
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i] < lines[j]
	})
	for _, line := range lines {
		r += fmt.Sprintf("\t\t%v\n", line)
	}

	if mermaid {
		r += "\tend\n"
	} else {
		r += "\t}\n"
	}
	return r, globalLines
}

func sortedHosts(hosts map[netip.Addr]*HostInfo) []netip.Addr {
	keys := make([]netip.Addr, 0, len(hosts))
	for key := range hosts {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].Compare(keys[j]) < 0
	})

	return keys
}

func sortedIndexes[V any](indexes map[uint32]V) []uint32 {
	keys := make([]uint32, 0, len(indexes))
	for key := range indexes {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	return keys
}
