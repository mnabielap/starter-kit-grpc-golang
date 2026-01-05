import sys
import os
import time
import grpc
import concurrent.futures

# Import path setup
sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import load_config, GRPC_HOST, GRPC_PORT

# Import generated protos
from api.proto.v1 import user_pb2
from api.proto.v1 import user_pb2_grpc

# --- CONFIGURATION ---
DURATION_SECONDS = 10   # How long to run the test
CONCURRENCY = 20        # Number of concurrent threads (simulated users)

# --- COLORS ---
class Colors:
    HEADER = '\033[95m'
    OKGREEN = '\033[92m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def run_benchmark():
    # 1. Setup Connection
    print(f"{Colors.HEADER}=== gRPC RPS BENCHMARK ==={Colors.ENDC}")
    print(f"Target: {GRPC_HOST}:{GRPC_PORT}")
    print(f"Duration: {DURATION_SECONDS}s | Threads: {CONCURRENCY}")

    channel = grpc.insecure_channel(f'{GRPC_HOST}:{GRPC_PORT}')
    stub = user_pb2_grpc.UserServiceStub(channel)

    # 2. Setup Auth
    token = load_config("accessToken")
    if not token:
        print(f"{Colors.FAIL}Error: No access token found. Run A2.auth_login.py first.{Colors.ENDC}")
        return
    
    metadata = [('authorization', f'Bearer {token}')]

    # 3. Prepare Request Object (Construct once to save CPU)
    # We use the raw Proto object, not a dict
    request = user_pb2.ListUsersRequest(
        page=1, 
        limit=10, 
        sort="created_at:desc"
    )

    # 4. Worker Function
    # This runs inside every thread
    def worker(end_time):
        count = 0
        errors = 0
        while time.time() < end_time:
            try:
                # Direct gRPC call (No logging overhead)
                stub.ListUsers(request, metadata=metadata)
                count += 1
            except grpc.RpcError:
                errors += 1
        return count, errors

    # 5. Execute Load Test
    print(f"\n{Colors.BOLD}Starting load test...{Colors.ENDC}")
    start_time = time.time()
    end_time = start_time + DURATION_SECONDS
    
    total_requests = 0
    total_errors = 0

    with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        # Launch threads
        futures = [executor.submit(worker, end_time) for _ in range(CONCURRENCY)]
        
        # Wait for all threads to finish
        for future in concurrent.futures.as_completed(futures):
            c, e = future.result()
            total_requests += c
            total_errors += e

    actual_duration = time.time() - start_time
    rps = total_requests / actual_duration

    # 6. Results
    print(f"\n{Colors.HEADER}=== RESULTS ==={Colors.ENDC}")
    print(f"Total Requests  : {total_requests}")
    print(f"Total Errors    : {total_errors}")
    print(f"Actual Duration : {actual_duration:.2f}s")
    print(f"{Colors.OKGREEN}{Colors.BOLD}RPS             : {rps:.2f} req/s{Colors.ENDC}")

    if total_errors > 0:
        print(f"{Colors.FAIL}Warning: {total_errors} requests failed.{Colors.ENDC}")

if __name__ == "__main__":
    run_benchmark()