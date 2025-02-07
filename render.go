package nebula

import (
	"fmt"
	"net/netip"
	"sort"
	"strings"
)

type edge struct {
	from string
	to   string
	dual bool
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
		r += "\tcompound=true\n"
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
				r += fmt.Sprintf("\t%v -> %v [dir=both]\n", line.from, line.to)
			} else {
				r += fmt.Sprintf("\t%v -> %v\n", line.from, line.to)
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

	crt := f.pki.getCertState().GetDefaultCertificate()
	clusterName := strings.Trim(crt.Name(), " ")
	clusterVpnIp := crt.Networks()[0].Addr()
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
		r += fmt.Sprintf("\t\tsubgraph %s.hosts[\"Hosts (vpn ip to index)\"]\n", clusterName)
	} else {
		r += fmt.Sprintf("\t\tsubgraph cluster_%s_hosts {\n", clusterName)
		r += "\t\t\tlabel=\"Hosts (vpn ip to index)\"\n"
	}
	hosts := sortedHosts(f.hostMap.Hosts)
	for _, vpnIp := range hosts {
		hi := f.hostMap.Hosts[vpnIp]
		if mermaid {
			r += fmt.Sprintf("\t\t\t%v.%v[\"%v\"]\n", clusterName, vpnIp, vpnIp)
			lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, vpnIp, clusterName, hi.localIndexId))
		} else {
			r += fmt.Sprintf("\t\t\t\"%v_%v\" [label=\"%v\"]\n", clusterName, vpnIp, vpnIp)
			lines = append(lines, fmt.Sprintf("\"%v_%v\" -> \"%v_%v\"", clusterName, vpnIp, clusterName, hi.localIndexId))
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
			r += fmt.Sprintf("\t\tsubgraph %s.relays[\"Relays (relay index to hostinfo)\"]\n", clusterName)
		} else {
			r += fmt.Sprintf("\t\tsubgraph cluster_%s_relays {\n", clusterName)
			r += "\t\t\tlabel=\"Relays (relay index to hostinfo)\"\n"
		}
		indexes := sortedIndexes(f.hostMap.Relays)
		for _, relayIndex := range indexes {
			hi := f.hostMap.Relays[relayIndex]
			if mermaid {
				r += fmt.Sprintf("\t\t\t%v.%v[\"%v (%v)\"]\n", clusterName, relayIndex, relayIndex, hi.vpnAddrs)
				lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, relayIndex, clusterName, hi.localIndexId))
			} else {
				r += fmt.Sprintf("\t\t\t\"%v_%v\" [label=\"%v (%v)\"]\n", clusterName, relayIndex, relayIndex, hi.vpnAddrs)
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
		r += fmt.Sprintf("\t\tsubgraph indexes.%s[\"Indexes (index to hostinfo)\"]\n", clusterName)
	} else {
		r += fmt.Sprintf("\t\tsubgraph cluster_%s_indexes {\n", clusterName)
		r += "\t\t\tlabel=\"Indexes (index to hostinfo)\"\n"
	}
	indexes := sortedIndexes(f.hostMap.Indexes)
	for _, idx := range indexes {
		hi, ok := f.hostMap.Indexes[idx]
		if ok {
			remoteClusterName := strings.Trim(hi.GetCert().Certificate.Name(), " ")
			if mermaid {
				r += fmt.Sprintf("\t\t\t%v.%v[\"%v (%v)\"]\n", clusterName, idx, idx, hi.remote)
				globalLines = append(globalLines, &edge{
					from: fmt.Sprintf("%v.%v", clusterName, idx),
					to:   fmt.Sprintf("%v.%v", remoteClusterName, hi.remoteIndexId),
				})
			} else {
				r += fmt.Sprintf("\t\t\t\"%v_%v\" [label=\"%v (%v)\"]\n", clusterName, idx, idx, hi.remote)
				globalLines = append(globalLines, &edge{
					from: fmt.Sprintf("\"%v_%v\"", clusterName, idx),
					to:   fmt.Sprintf("\"%v_%v\"", remoteClusterName, hi.remoteIndexId),
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
