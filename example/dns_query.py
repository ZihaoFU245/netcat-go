import subprocess
import struct
import sys

def parse_dns_response(data):
    if len(data) < 12:
        print("Response too short to be DNS packet")
        return

    # Parse Header
    trans_id, flags, qdcount, ancount, nscount, arcount = struct.unpack('!HHHHHH', data[:12])
    print(f"--- DNS Response ---")
    print(f"Transaction ID: {hex(trans_id)}")
    print(f"Flags: {hex(flags)}")
    print(f"Questions: {qdcount}")
    print(f"Answers: {ancount}")
    
    offset = 12
    
    # Skip Questions section
    for _ in range(qdcount):
        while True:
            if offset >= len(data): return
            length = data[offset]
            offset += 1
            if length == 0:
                break
            if length & 0xC0 == 0xC0: # Pointer
                offset += 1
                break
            offset += length
        offset += 4 # Skip Type (2) and Class (2)

    # Parse Answers section
    print("--- Answers ---")
    for _ in range(ancount):
        if offset >= len(data): break
        
        # Name
        if data[offset] & 0xC0 == 0xC0:
            offset += 2
        else:
             while True:
                length = data[offset]
                offset += 1
                if length == 0:
                    break
                offset += length
        
        # Type, Class, TTL, RDLENGTH
        type_, class_, ttl, rdlength = struct.unpack('!HHIH', data[offset:offset+10])
        offset += 10
        
        rdata = data[offset:offset+rdlength]
        
        if type_ == 1: # A Record (IPv4)
            if len(rdata) == 4:
                ip = struct.unpack('!BBBB', rdata)
                print(f"Type A (IPv4): {'.'.join(map(str, ip))}")
            else:
                print(f"Type A: (Invalid length)")
        elif type_ == 28: # AAAA Record (IPv6)
            print(f"Type AAAA (IPv6): {rdata.hex()}")
        else:
            print(f"Type {type_}: Data length {rdlength}")
            
        offset += rdlength

def main():
    # DNS Query for example.com (Type A)
    query = b'\xAA\xAA\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00\x07example\x03com\x00\x00\x01\x00\x01'
    
    print("Sending DNS query via nc.exe...")
    
    try:
        # Run nc.exe with UDP mode and 2s timeout
        # Ensure nc.exe is built and in the current directory
        process = subprocess.Popen(
            ['./nc.exe','-i', '3', '-u', '8.8.8.8', '53'],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        
        stdout, stderr = process.communicate(input=query)
        
        if stderr:
            # nc prints errors (like timeout) to stderr
            print(f"nc stderr: {stderr.decode('utf-8', errors='ignore').strip()}")

        if stdout:
            print(f"Received {len(stdout)} bytes.")
            parse_dns_response(stdout)
        else:
            print("No data received.")

    except FileNotFoundError:
        print("Error: nc.exe not found. Please build it first using 'go build -o nc.exe main.go'")
    except Exception as e:
        print(f"An error occurred: {e}")

if __name__ == "__main__":
    main()
