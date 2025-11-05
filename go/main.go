// DBViewerServer (Go) - 单文件静态服务器
// 构建：go build -ldflags="-s -w" -o DBViewerServer.exe

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	// 解析参数
	dir := getArg("dir", exeDir())
	entry := getArg("entry", findEntry(dir))
	port := getArg("port", "58490")

	// 静态文件 + 禁缓存
	fs := http.FileServer(http.Dir(dir))
	handler := noCache(fs)

	// 默认入口：如果直接访问 / 则重定向到 entry（若有）
	if entry != "" {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/"+entry, http.StatusFound)
				return
			}
			handler.ServeHTTP(w, r)
		})
	} else {
		http.Handle("/", handler)
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	addr := "0.0.0.0:" + port

	// 打印横幅
	fmt.Println()
	fmt.Println("████████████████████  DBViewerServer (Go)  ██████████████████")
	fmt.Println(" Dir   :", dir)
	fmt.Println(" Entry :", ifEmpty(entry, "(none)"))
	fmt.Println(" Port  :", port)
	fmt.Println(" Local :", "http://127.0.0.1:"+port+"/"+entry)
	for _, ip := range localIPv4() {
		fmt.Println(" Share :", "http://"+ip+":"+port+"/"+entry)
	}
	fmt.Println(" NOTE  : 保持窗口开启；同网段其它设备访问 Share 地址")
	fmt.Println("██████████████████████████████████████████████████████████")
	fmt.Println()

	_ = allowFirewall(port) // 尝试放行防火墙（忽略失败）

	// 延迟打开本机页面
	go func() {
		time.Sleep(300 * time.Millisecond)
		openURL("http://127.0.0.1:" + port + "/" + entry)
	}()

	// 真正启动
	server := &http.Server{
		Addr:         addr,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "-1")
		next.ServeHTTP(w, r)
	})
}

func exeDir() string {
	p, _ := os.Executable()
	return filepath.Dir(p)
}

func findEntry(dir string) string {
	for _, n := range []string{"viewer_tool.html", "viewer_local.html", "index.html"} {
		if _, err := os.Stat(filepath.Join(dir, n)); err == nil {
			return n
		}
	}
	return ""
}

func getArg(key, def string) string {
	prefix := "--" + key + "="
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, prefix) {
			return strings.TrimPrefix(a, prefix)
		}
	}
	return def
}

func localIPv4() []string {
	var out []string
	ifaces, _ := net.Interfaces()
	for _, ifc := range ifaces {
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				ip := ipnet.IP
				if ip.To4() != nil && !ip.IsLoopback() && !strings.HasPrefix(ip.String(), "169.254.") {
					out = append(out, ip.String())
				}
			}
		}
	}
	return out
}

func openURL(u string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	case "darwin":
		_ = exec.Command("open", u).Start()
	default:
		_ = exec.Command("xdg-open", u).Start()
	}
}

func allowFirewall(port string) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=DBViewer "+port, "dir=in", "action=allow", "protocol=TCP", "localport="+port)
	return cmd.Run()
}

func ifEmpty(v, alt string) string {
	if v == "" {
		return alt
	}
	return v
}
