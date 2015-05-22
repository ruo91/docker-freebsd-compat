package portallocator

import (
	"net"
	"testing"
)

func TestRequestNewPort(t *testing.T) {
	p := New()

	port, err := p.RequestPort(defaultIP, "tcp", 0)
	if err != nil {
		t.Fatal(err)
	}

	if expected := p.Begin; port != expected {
		t.Fatalf("Expected port %d got %d", expected, port)
	}
}

func TestRequestSpecificPort(t *testing.T) {
	p := New()

	port, err := p.RequestPort(defaultIP, "tcp", 5000)
	if err != nil {
		t.Fatal(err)
	}
	if port != 5000 {
		t.Fatalf("Expected port 5000 got %d", port)
	}
}

func TestReleasePort(t *testing.T) {
	p := New()

	port, err := p.RequestPort(defaultIP, "tcp", 5000)
	if err != nil {
		t.Fatal(err)
	}
	if port != 5000 {
		t.Fatalf("Expected port 5000 got %d", port)
	}

	if err := p.ReleasePort(defaultIP, "tcp", 5000); err != nil {
		t.Fatal(err)
	}
}

func TestReuseReleasedPort(t *testing.T) {
	p := New()

	port, err := p.RequestPort(defaultIP, "tcp", 5000)
	if err != nil {
		t.Fatal(err)
	}
	if port != 5000 {
		t.Fatalf("Expected port 5000 got %d", port)
	}

	if err := p.ReleasePort(defaultIP, "tcp", 5000); err != nil {
		t.Fatal(err)
	}

	port, err = p.RequestPort(defaultIP, "tcp", 5000)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReleaseUnreadledPort(t *testing.T) {
	p := New()

	port, err := p.RequestPort(defaultIP, "tcp", 5000)
	if err != nil {
		t.Fatal(err)
	}
	if port != 5000 {
		t.Fatalf("Expected port 5000 got %d", port)
	}

	port, err = p.RequestPort(defaultIP, "tcp", 5000)

	switch err.(type) {
	case ErrPortAlreadyAllocated:
	default:
		t.Fatalf("Expected port allocation error got %s", err)
	}
}

func TestUnknowProtocol(t *testing.T) {
	if _, err := New().RequestPort(defaultIP, "tcpp", 0); err != ErrUnknownProtocol {
		t.Fatalf("Expected error %s got %s", ErrUnknownProtocol, err)
	}
}

func TestAllocateAllPorts(t *testing.T) {
	p := New()

	for i := 0; i <= p.End-p.Begin; i++ {
		port, err := p.RequestPort(defaultIP, "tcp", 0)
		if err != nil {
			t.Fatal(err)
		}

		if expected := p.Begin + i; port != expected {
			t.Fatalf("Expected port %d got %d", expected, port)
		}
	}

	if _, err := p.RequestPort(defaultIP, "tcp", 0); err != ErrAllPortsAllocated {
		t.Fatalf("Expected error %s got %s", ErrAllPortsAllocated, err)
	}

	_, err := p.RequestPort(defaultIP, "udp", 0)
	if err != nil {
		t.Fatal(err)
	}

	// release a port in the middle and ensure we get another tcp port
	port := p.Begin + 5
	if err := p.ReleasePort(defaultIP, "tcp", port); err != nil {
		t.Fatal(err)
	}
	newPort, err := p.RequestPort(defaultIP, "tcp", 0)
	if err != nil {
		t.Fatal(err)
	}
	if newPort != port {
		t.Fatalf("Expected port %d got %d", port, newPort)
	}

	// now pm.last == newPort, release it so that it's the only free port of
	// the range, and ensure we get it back
	if err := p.ReleasePort(defaultIP, "tcp", newPort); err != nil {
		t.Fatal(err)
	}
	port, err = p.RequestPort(defaultIP, "tcp", 0)
	if err != nil {
		t.Fatal(err)
	}
	if newPort != port {
		t.Fatalf("Expected port %d got %d", newPort, port)
	}
}

func BenchmarkAllocatePorts(b *testing.B) {
	p := New()

	for i := 0; i < b.N; i++ {
		for i := 0; i <= p.End-p.Begin; i++ {
			port, err := p.RequestPort(defaultIP, "tcp", 0)
			if err != nil {
				b.Fatal(err)
			}

			if expected := p.Begin + i; port != expected {
				b.Fatalf("Expected port %d got %d", expected, port)
			}
		}
		p.ReleaseAll()
	}
}

func TestPortAllocation(t *testing.T) {
	p := New()

	ip := net.ParseIP("192.168.0.1")
	ip2 := net.ParseIP("192.168.0.2")
	if port, err := p.RequestPort(ip, "tcp", 80); err != nil {
		t.Fatal(err)
	} else if port != 80 {
		t.Fatalf("Acquire(80) should return 80, not %d", port)
	}
	port, err := p.RequestPort(ip, "tcp", 0)
	if err != nil {
		t.Fatal(err)
	}
	if port <= 0 {
		t.Fatalf("Acquire(0) should return a non-zero port")
	}

	if _, err := p.RequestPort(ip, "tcp", port); err == nil {
		t.Fatalf("Acquiring a port already in use should return an error")
	}

	if newPort, err := p.RequestPort(ip, "tcp", 0); err != nil {
		t.Fatal(err)
	} else if newPort == port {
		t.Fatalf("Acquire(0) allocated the same port twice: %d", port)
	}

	if _, err := p.RequestPort(ip, "tcp", 80); err == nil {
		t.Fatalf("Acquiring a port already in use should return an error")
	}
	if _, err := p.RequestPort(ip2, "tcp", 80); err != nil {
		t.Fatalf("It should be possible to allocate the same port on a different interface")
	}
	if _, err := p.RequestPort(ip2, "tcp", 80); err == nil {
		t.Fatalf("Acquiring a port already in use should return an error")
	}
	if err := p.ReleasePort(ip, "tcp", 80); err != nil {
		t.Fatal(err)
	}
	if _, err := p.RequestPort(ip, "tcp", 80); err != nil {
		t.Fatal(err)
	}

	port, err = p.RequestPort(ip, "tcp", 0)
	if err != nil {
		t.Fatal(err)
	}
	port2, err := p.RequestPort(ip, "tcp", port+1)
	if err != nil {
		t.Fatal(err)
	}
	port3, err := p.RequestPort(ip, "tcp", 0)
	if err != nil {
		t.Fatal(err)
	}
	if port3 == port2 {
		t.Fatal("Requesting a dynamic port should never allocate a used port")
	}
}

func TestNoDuplicateBPR(t *testing.T) {
	p := New()

	if port, err := p.RequestPort(defaultIP, "tcp", p.Begin); err != nil {
		t.Fatal(err)
	} else if port != p.Begin {
		t.Fatalf("Expected port %d got %d", p.Begin, port)
	}

	if port, err := p.RequestPort(defaultIP, "tcp", 0); err != nil {
		t.Fatal(err)
	} else if port == p.Begin {
		t.Fatalf("Acquire(0) allocated the same port twice: %d", port)
	}
}
