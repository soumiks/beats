package beat

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/gosigar"
)

type SystemLoad struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type CpuTimes struct {
	User          uint64  `json:"user"`
	UserPercent   float64 `json:"user_p"`
	Nice          uint64  `json:"nice"`
	System        uint64  `json:"system"`
	SystemPercent float64 `json:"system_p"`
	Idle          uint64  `json:"idle"`
	IOWait        uint64  `json:"iowait"`
	Irq           uint64  `json:"irq"`
	SoftIrq       uint64  `json:"softirq"`
	Steal         uint64  `json:"steal"`
}

type MemStat struct {
	Total             uint64  `json:"total"`
	Used              uint64  `json:"used"`
	Free              uint64  `json:"free"`
	UsedPercent       float64 `json:"used_p"`
	ActualUsed        uint64  `json:"actual_used"`
	ActualFree        uint64  `json:"actual_free"`
	ActualUsedPercent float64 `json:"actual_used_p"`
}

type ProcMemStat struct {
	Size       uint64  `json:"size"`
	Rss        uint64  `json:"rss"`
	RssPercent float64 `json:"rss_p"`
	Share      uint64  `json:"share"`
}

type ProcCpuTime struct {
	User        uint64  `json:"user"`
	UserPercent float64 `json:"user_p"`
	System      uint64  `json:"system"`
	Total       uint64  `json:"total"`
	Start       string  `json:"start_time"`
}

type ProcConnection struct {
	LocalIp    string `json:"localip"`
	LocalPort  string `json:"localport"`
	RemoteIp   string `json:"remoteip"`
	RemotePort string `json:"remoteport"`
}

type ProcConnections struct {
	Connections []ProcConnection `json:"connections"`
}

type Process struct {
	Pid         int              `json:"pid"`
	Ppid        int              `json:"ppid"`
	Name        string           `json:"name"`
	State       string           `json:"state"`
	Mem         *ProcMemStat     `json:"mem"`
	Cpu         *ProcCpuTime     `json:"cpu"`
	Connections *ProcConnections `json:"connections"`
	ctime       time.Time
}

type FileSystemStat struct {
	DevName     string  `json:"device_name"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_p"`
	Free        uint64  `json:"free"`
	Avail       uint64  `json:"avail"`
	Files       uint64  `json:"files"`
	FreeFiles   uint64  `json:"free_files"`
	Mount       string  `json:"mount_point"`
	ctime       time.Time
}

func (f *FileSystemStat) String() string {

	return fmt.Sprintf("device name: %s, total: %d, used %d, used pct %.2f, free: %d, avail: %d, files: %d, free files: %d, mount: %s",
		f.DevName, f.Total, f.Used, f.UsedPercent, f.Free, f.Avail, f.Files, f.FreeFiles, f.Mount)
}

func (p *Process) String() string {

	return fmt.Sprintf("pid: %d, ppid: %d, name: %s, state: %s, mem: %s, cpu: %s, connections: %s",
		p.Pid, p.Ppid, p.Name, p.State, p.Mem.String(), p.Cpu.String(), p.Connections.String())
}

func (m *ProcMemStat) String() string {

	return fmt.Sprintf("%d size, %d rss, %d share", m.Size, m.Rss, m.Share)
}

func (t *ProcCpuTime) String() string {
	return fmt.Sprintf("started at %s, %d total %.2f%%CPU, %d us, %d sys", t.Start, t.Total, t.UserPercent, t.User, t.System)

}

func (c *ProcConnections) String() string {
	ret := ""
	for _, conn := range c.Connections {
		ret += fmt.Sprintf("%d localip, %d localport, %d remoteip, %d remoteport", conn.LocalIp, conn.LocalPort, conn.RemoteIp, conn.RemotePort)
	}
	return ret

}

func (m *MemStat) String() string {

	return fmt.Sprintf("%d total, %d used, %d actual used, %d free, %d actual free", m.Total, m.Used, m.ActualUsed,
		m.Free, m.ActualFree)
}

func (t *SystemLoad) String() string {

	return fmt.Sprintf("%.2f %.2f %.2f", t.Load1, t.Load5, t.Load15)
}

func (t *CpuTimes) String() string {

	return fmt.Sprintf("%d user, %d system, %d nice, %d iddle, %d iowait, %d irq, %d softirq, %d steal",
		t.User, t.System, t.Nice, t.Idle, t.IOWait, t.Irq, t.SoftIrq, t.Steal)

}
func GetSystemLoad() (*SystemLoad, error) {

	concreteSigar := sigar.ConcreteSigar{}
	avg, err := concreteSigar.GetLoadAverage()
	if err != nil {
		return nil, err
	}

	return &SystemLoad{
		Load1:  avg.One,
		Load5:  avg.Five,
		Load15: avg.Fifteen,
	}, nil
}

func GetCpuTimes() (*CpuTimes, error) {

	cpu := sigar.Cpu{}
	err := cpu.Get()
	if err != nil {
		return nil, err
	}

	return &CpuTimes{
		User:    cpu.User,
		Nice:    cpu.Nice,
		System:  cpu.Sys,
		Idle:    cpu.Idle,
		IOWait:  cpu.Wait,
		Irq:     cpu.Irq,
		SoftIrq: cpu.SoftIrq,
		Steal:   cpu.Stolen,
	}, nil
}

func GetCpuTimesList() ([]CpuTimes, error) {

	cpuList := sigar.CpuList{}
	err := cpuList.Get()
	if err != nil {
		return nil, err
	}

	cpuTimes := make([]CpuTimes, len(cpuList.List))

	for i, cpu := range cpuList.List {
		cpuTimes[i] = CpuTimes{
			User:    cpu.User,
			Nice:    cpu.Nice,
			System:  cpu.Sys,
			Idle:    cpu.Idle,
			IOWait:  cpu.Wait,
			Irq:     cpu.Irq,
			SoftIrq: cpu.SoftIrq,
			Steal:   cpu.Stolen,
		}
	}

	return cpuTimes, nil
}

func GetMemory() (*MemStat, error) {

	mem := sigar.Mem{}
	err := mem.Get()
	if err != nil {
		return nil, err
	}

	return &MemStat{
		Total:      mem.Total,
		Used:       mem.Used,
		Free:       mem.Free,
		ActualFree: mem.ActualFree,
		ActualUsed: mem.ActualUsed,
	}, nil
}

func GetSwap() (*MemStat, error) {

	swap := sigar.Swap{}
	err := swap.Get()
	if err != nil {
		return nil, err
	}

	return &MemStat{
		Total: swap.Total,
		Used:  swap.Used,
		Free:  swap.Free,
	}, nil

}

func Pids() ([]int, error) {

	pids := sigar.ProcList{}
	err := pids.Get()
	if err != nil {
		return nil, err
	}
	return pids.List, nil
}

func getProcState(b byte) string {

	switch b {
	case 'S':
		return "sleeping"
	case 'R':
		return "running"
	case 'D':
		return "idle"
	case 'T':
		return "stopped"
	case 'Z':
		return "zombie"
	}
	return "unknown"
}

func GetProcess(pid int) (*Process, error) {

	state := sigar.ProcState{}
	mem := sigar.ProcMem{}
	cpu := sigar.ProcTime{}

	err := state.Get(pid)
	if err != nil {
		return nil, fmt.Errorf("Error getting state info: %v", err)
	}

	err = mem.Get(pid)
	if err != nil {
		return nil, fmt.Errorf("Error getting mem info: %v", err)
	}

	err = cpu.Get(pid)
	if err != nil {
		return nil, fmt.Errorf("Error getting cpu info: %v", err)
	}

	proc := Process{
		Pid:   pid,
		Ppid:  state.Ppid,
		Name:  state.Name,
		State: getProcState(byte(state.State)),
		Mem: &ProcMemStat{
			Size:  mem.Size,
			Rss:   mem.Resident,
			Share: mem.Share,
		},
		Cpu: &ProcCpuTime{
			Start:  cpu.FormatStartTime(),
			Total:  cpu.Total,
			User:   cpu.User,
			System: cpu.Sys,
		},
		Connections: &ProcConnections{
			Connections: nil,
		},
	}
	proc.ctime = time.Now()

	return &proc, nil
}

func GetFileSystemList() ([]sigar.FileSystem, error) {

	fss := sigar.FileSystemList{}
	err := fss.Get()
	if err != nil {
		return nil, err
	}

	return fss.List, nil
}

func GetFileSystemStat(fs sigar.FileSystem) (*FileSystemStat, error) {

	stat := sigar.FileSystemUsage{}
	err := stat.Get(fs.DirName)
	if err != nil {
		return nil, err
	}

	filesystem := FileSystemStat{
		DevName:   fs.DevName,
		Total:     stat.Total,
		Free:      stat.Free,
		Avail:     stat.Avail,
		Used:      stat.Used,
		Files:     stat.Files,
		FreeFiles: stat.FreeFiles,
		Mount:     fs.DirName,
	}

	return &filesystem, nil
}

func GetProcessTCPConnections(pid int) []ProcConnection {

	filename := "/proc/" + strconv.Itoa(pid) + "/net/tcp"
	logp.Debug("topbeat", "GetProcessTCPConnections %s\n", filename)

	connection_array := []ProcConnection{}
	dat, _ := ioutil.ReadFile(filename)

	lines := strings.Split(string(dat), "\n")

	for _, line := range lines[1:] {

		// split by space
		columns := strings.Split(line, " ")

		if len(columns) > 5 {
			tempconnection := ProcConnection{"", "", "", ""}
			local_ip_port := strings.Split(columns[4], ":")
			if len(local_ip_port) > 0 {
				tempconnection.LocalIp = local_ip_port[0]
			}
			if len(local_ip_port) > 1 {
				tempconnection.LocalPort = local_ip_port[1]
			}
			remote_ip_port := strings.Split(columns[5], ":")
			if len(remote_ip_port) > 0 {
				tempconnection.RemoteIp = remote_ip_port[0]
			}
			if len(remote_ip_port) > 1 {
				tempconnection.RemotePort = remote_ip_port[1]
			}

			connection_array = append(connection_array, tempconnection)
		}
	}

	return connection_array
}