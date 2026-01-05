import sys
import os
import time
import grpc
import requests
import concurrent.futures

# Import path setup
sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import load_config, GRPC_HOST, GRPC_PORT

# Import generated protos
from api.proto.v1 import user_pb2
from api.proto.v1 import user_pb2_grpc

# --- CONFIGURATION ---
REST_HOST = "localhost"
REST_PORT = "8080"
DURATION_SECONDS = 10   # Duration for EACH test
CONCURRENCY = 30        # Number of concurrent threads

# --- COLORS ---
class Colors:
    HEADER = '\033[95m'
    OKGREEN = '\033[92m'
    OKCYAN = '\033[96m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def test_grpc(token):
    print(f"\n{Colors.OKCYAN}--- Testing gRPC (Port {GRPC_PORT}) ---{Colors.ENDC}")
    
    # Setup
    channel = grpc.insecure_channel(f'{GRPC_HOST}:{GRPC_PORT}')
    stub = user_pb2_grpc.UserServiceStub(channel)
    metadata = [('authorization', f'Bearer {token}')]
    
    # Pre-construct request
    request = user_pb2.ListUsersRequest(page=1, limit=10, sort="created_at:desc")

    def grpc_worker(end_time):
        count = 0
        errors = 0
        while time.time() < end_time:
            try:
                stub.ListUsers(request, metadata=metadata)
                count += 1
            except grpc.RpcError:
                errors += 1
        return count, errors

    # Run Benchmark
    start_time = time.time()
    end_time = start_time + DURATION_SECONDS
    
    total_req = 0
    total_err = 0

    with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        futures = [executor.submit(grpc_worker, end_time) for _ in range(CONCURRENCY)]
        for future in concurrent.futures.as_completed(futures):
            c, e = future.result()
            total_req += c
            total_err += e

    duration = time.time() - start_time
    rps = total_req / duration
    print(f"Requests: {total_req} | Errors: {total_err} | Time: {duration:.2f}s")
    return rps

def test_rest(token):
    print(f"\n{Colors.OKCYAN}--- Testing REST/Gateway (Port {REST_PORT}) ---{Colors.ENDC}")
    
    url = f"http://{REST_HOST}:{REST_PORT}/v1/users"
    # Query params to match gRPC request
    params = {"page": 1, "limit": 10, "sort": "created_at:desc"}
    headers = {"Authorization": f"Bearer {token}"}

    def rest_worker(end_time):
        count = 0
        errors = 0
        # Use Session for Connection Pooling (Keep-Alive)
        # This makes the comparison fair, otherwise HTTP handshake kills performance
        session = requests.Session()
        session.headers.update(headers)
        
        while time.time() < end_time:
            try:
                resp = session.get(url, params=params)
                if resp.status_code == 200:
                    count += 1
                else:
                    errors += 1
            except Exception:
                errors += 1
        return count, errors

    # Run Benchmark
    start_time = time.time()
    end_time = start_time + DURATION_SECONDS
    
    total_req = 0
    total_err = 0

    with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        futures = [executor.submit(rest_worker, end_time) for _ in range(CONCURRENCY)]
        for future in concurrent.futures.as_completed(futures):
            c, e = future.result()
            total_req += c
            total_err += e

    duration = time.time() - start_time
    rps = total_req / duration
    print(f"Requests: {total_req} | Errors: {total_err} | Time: {duration:.2f}s")
    return rps

def main():
    print(f"{Colors.HEADER}{Colors.BOLD}=== BENCHMARK: gRPC vs REST Gateway ==={Colors.ENDC}")
    print(f"Duration: {DURATION_SECONDS}s per test | Threads: {CONCURRENCY}")

    # Load Token
    token = load_config("accessToken")
    if not token:
        print(f"{Colors.FAIL}Error: No access token found. Run A2.auth_login.py first.{Colors.ENDC}")
        return

    # Run Tests
    grpc_rps = test_grpc(token)
    time.sleep(1) # Cool down
    rest_rps = test_rest(token)

    # Calculate Comparison
    print(f"\n{Colors.HEADER}=== FINAL RESULTS ==={Colors.ENDC}")
    print(f"{Colors.BOLD}gRPC RPS :{Colors.ENDC} {grpc_rps:,.2f}")
    print(f"{Colors.BOLD}REST RPS :{Colors.ENDC} {rest_rps:,.2f}")

    if rest_rps > 0:
        multiplier = grpc_rps / rest_rps
        print(f"\n{Colors.OKGREEN}>> gRPC is {multiplier:.2f}x faster than REST in this test.{Colors.ENDC}")
    
    print(f"\n{Colors.HEADER}Note: Real world performance gap is often larger because Python threads limit max gRPC throughput here.{Colors.ENDC}")

if __name__ == "__main__":
    main()