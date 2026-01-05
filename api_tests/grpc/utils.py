import json
import os
import time
import grpc
from google.protobuf.json_format import MessageToDict, ParseDict

# --- CONFIGURATION constants ---
# gRPC usually runs on a different port (e.g., 50051) than REST
GRPC_HOST = "localhost"
GRPC_PORT = "50051"
CONFIG_FILE_BASE = "secrets.json"

# --- HELPER: Config Management (Secrets) ---

def save_config(key, value):
    """Saves a key-value pair to secrets.json for token persistence."""
    global CONFIG_FILE_BASE
    CONFIG_FILE = os.path.join(os.path.dirname(__file__), CONFIG_FILE_BASE)
    data = {}
    if os.path.exists(CONFIG_FILE):
        try:
            with open(CONFIG_FILE, 'r') as f:
                data = json.load(f)
        except:
            pass
    data[key] = value
    with open(CONFIG_FILE, 'w') as f:
        json.dump(data, f, indent=4)

def load_config(key):
    """Loads a value from secrets.json."""
    global CONFIG_FILE_BASE
    CONFIG_FILE = os.path.join(os.path.dirname(__file__), CONFIG_FILE_BASE)
    if not os.path.exists(CONFIG_FILE):
        return None
    try:
        with open(CONFIG_FILE, 'r') as f:
            data = json.load(f)
        return data.get(key)
    except:
        return None

# --- HELPER: Compatibility Wrapper ---

class GrpcResponseProxy:
    """
    Wraps the gRPC result (or error) to mimic a response object.
    Allows easy access to .json() and status codes.
    """
    def __init__(self, result_dict, grpc_code, grpc_details):
        self.result_dict = result_dict
        self.grpc_code = grpc_code      # e.g., grpc.StatusCode.OK
        self.grpc_details = grpc_details # Error message string if any
        
        # Helper: Map grpc.StatusCode.OK to a boolean or mimic HTTP 200/500 logic broadly
        self.ok = (grpc_code == grpc.StatusCode.OK)

    def json(self):
        """Returns the parsed dictionary body."""
        return self.result_dict

    @property
    def status_code(self):
        """Returns the string name of the status code (e.g., 'OK', 'NOT_FOUND')."""
        return self.grpc_code.name if self.grpc_code else "UNKNOWN"

# --- MAIN FUNCTION ---

def run_grpc_test(stub,
                  method_name: str,
                  request_proto_class,
                  payload: dict = None,
                  metadata: list = None,
                  output_file: str = "response.json",
                  print_pretty_response: bool = True):
    """
    Executes a gRPC call, prints details, and saves results to a file.
    
    Args:
        stub: The initialized gRPC Stub object (e.g., auth_pb2_grpc.AuthServiceStub).
        method_name (str): The name of the method to call on the stub (e.g., 'Login').
        request_proto_class: The Protobuf class for the request (e.g., auth_pb2.LoginRequest).
        payload (dict): The data to send (will be converted to Proto).
        metadata (list): List of tuples for headers, e.g. [('authorization', 'Bearer ...')]
        output_file (str): Filename to save the JSON output.
    """
    
    output_file = os.path.join(os.path.dirname(__file__), output_file)
    timestamp_start = time.time()
    
    print("\n===== gRPC REQUEST START =====")
    print(f"Method: {method_name}")
    print(f"Target: {GRPC_HOST}:{GRPC_PORT}")

    # 1. Prepare Metadata (Headers)
    if metadata:
        print("Metadata (Headers):")
        # specific print for tuples
        meta_dict = {k: v for k, v in metadata} 
        print(json.dumps(meta_dict, indent=4, ensure_ascii=False))
    else:
        print("Metadata: <none>")

    # 2. Prepare Request Body
    if payload is None:
        payload = {}
    
    print("Request Payload (JSON input):")
    print(json.dumps(payload, indent=4, ensure_ascii=False))

    try:
        # Convert Dictionary -> Protobuf Message
        req_proto = request_proto_class()
        ParseDict(payload, req_proto)
        
        # 3. Send Request
        print("\nSending gRPC request...")
        rpc_method = getattr(stub, method_name)
        
        send_time = time.time()
        
        # Execute RPC
        response_proto = rpc_method(req_proto, metadata=metadata)
        
        receive_time = time.time()
        duration = receive_time - send_time
        print(f"Response received (duration {duration:.3f}s)")

        # 4. Handle Success
        # Convert Protobuf Message -> Dictionary
        resp_dict = MessageToDict(
            response_proto, 
            preserving_proto_field_name=True, # Use snake_case (e.g. access_token) instead of camelCase
            use_integers_for_enums=False
        )
        
        grpc_code = grpc.StatusCode.OK
        grpc_details = "OK"

        print("\n===== RESPONSE SUMMARY =====")
        print(f"Status: {grpc_code.name}")
        
        if print_pretty_response:
            print("\nResponse Body:")
            print(json.dumps(resp_dict, indent=4, ensure_ascii=False))
        else:
            print(f"\nResponse Body: {json.dumps(resp_dict)}")

        # Save to file
        result_summary = {
            "request": {
                "method": method_name,
                "payload": payload,
                "metadata": metadata
            },
            "response": {
                "status": grpc_code.name,
                "body": resp_dict,
                "duration": duration
            },
            "timestamp": time.time()
        }

    except grpc.RpcError as e:
        # 5. Handle gRPC Errors
        receive_time = time.time()
        duration = receive_time - send_time
        
        grpc_code = e.code()
        grpc_details = e.details()
        
        print("\n===== gRPC ERROR =====")
        print(f"Status Code: {grpc_code.name}")
        print(f"Details: {grpc_details}")
        
        # Try to parse trailing metadata if available
        # err_meta = e.trailing_metadata()
        
        resp_dict = {
            "error": grpc_code.name,
            "message": grpc_details
        }
        
        result_summary = {
            "request": {
                "method": method_name,
                "payload": payload
            },
            "error": {
                "status": grpc_code.name,
                "details": grpc_details,
                "duration": duration
            }
        }
        
    # Write output file
    try:
        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(result_summary, f, indent=4, ensure_ascii=False)
        print(f"\nFull result saved to {os.path.abspath(output_file)}")
    except Exception as ex:
        print(f"Failed to save result: {ex}")

    timestamp_end = time.time()
    print(f"Total elapsed time: {timestamp_end - timestamp_start:.3f}s")
    print("===== REQUEST END =====\n")

    return GrpcResponseProxy(resp_dict, grpc_code, grpc_details)