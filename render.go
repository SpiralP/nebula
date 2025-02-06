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
	return RenderHostmaps(interfaces...)
}

func RenderHostmaps(interfaces ...*Interface) string {
	var lines []*edge
	r := "graph TB\n"
	for _, c := range interfaces {
		sr, se := renderHostmap(c)
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

	for _, line := range lines {
		if line.dual {
			r += fmt.Sprintf("\t%v <--> %v\n", line.from, line.to)
		} else {
			r += fmt.Sprintf("\t%v --> %v\n", line.from, line.to)
		}

	}

	return r
}

func renderHostmap(f *Interface) (string, []*edge) {
	var lines []string
	var globalLines []*edge

	clusterName := strings.Trim(f.pki.GetCertState().Certificate.Details.Name, " ")
	clusterVpnIp := f.pki.GetCertState().Certificate.Details.Ips[0].IP
	r := fmt.Sprintf("\tsubgraph %s[\"%s (%s)\"]\n", clusterName, clusterName, clusterVpnIp)

	f.hostMap.RLock()
	defer f.hostMap.RUnlock()

	// Draw the vpn to index nodes
	r += fmt.Sprintf("\t\tsubgraph %s.hosts[\"Hosts (vpn ip to index)\"]\n", clusterName)
	hosts := sortedHosts(f.hostMap.Hosts)
	for _, vpnIp := range hosts {
		hi := f.hostMap.Hosts[vpnIp]
		r += fmt.Sprintf("\t\t\t%v.%v[\"%v\"]\n", clusterName, vpnIp, vpnIp)
		lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, vpnIp, clusterName, hi.localIndexId))

		for _, relayIp := range hi.relayState.CopyRelayIps() {
			lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, vpnIp, clusterName, relayIp))
		}

		for _, relayIp := range hi.relayState.CopyRelayForIdxs() {
			lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, vpnIp, clusterName, relayIp))
		}
	}
	r += "\t\tend\n"

	// Draw the relay hostinfos
	if len(f.hostMap.Relays) > 0 {
		r += fmt.Sprintf("\t\tsubgraph %s.relays[\"Relays (relay index to hostinfo)\"]\n", clusterName)
		for relayIndex, hi := range f.hostMap.Relays {
			r += fmt.Sprintf("\t\t\t%v.%v[\"%v\"]\n", clusterName, relayIndex, relayIndex)
			lines = append(lines, fmt.Sprintf("%v.%v --> %v.%v", clusterName, relayIndex, clusterName, hi.localIndexId))
		}
		r += "\t\tend\n"
	}

	// Draw the local index to relay or remote index nodes
	r += fmt.Sprintf("\t\tsubgraph indexes.%s[\"Indexes (index to hostinfo)\"]\n", clusterName)
	indexes := sortedIndexes(f.hostMap.Indexes)
	for _, idx := range indexes {
		hi, ok := f.hostMap.Indexes[idx]
		if ok {
			r += fmt.Sprintf("\t\t\t%v.%v[\"%v (%v)\"]\n", clusterName, idx, idx, hi.remote)
			remoteClusterName := strings.Trim(hi.GetCert().Details.Name, " ")
			globalLines = append(globalLines, &edge{from: fmt.Sprintf("%v.%v", clusterName, idx), to: fmt.Sprintf("%v.%v", remoteClusterName, hi.remoteIndexId)})
			_ = hi
		}
	}
	r += "\t\tend\n"

	// Add the edges inside this host
	for _, line := range lines {
		r += fmt.Sprintf("\t\t%v\n", line)
	}

	r += "\tend\n"
	return r, globalLines
}

func sortedHosts(hosts map[netip.Addr]*HostInfo) []netip.Addr {
	keys := make([]netip.Addr, 0, len(hosts))
	for key := range hosts {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].Compare(keys[j]) > 0
	})

	return keys
}

func sortedIndexes(indexes map[uint32]*HostInfo) []uint32 {
	keys := make([]uint32, 0, len(indexes))
	for key := range indexes {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	return keys
}
