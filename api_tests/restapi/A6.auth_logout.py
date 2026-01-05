import sys
import os
sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import send_and_print, BASE_URL, load_config

print("--- LOGOUT ---")

url = f"{BASE_URL}/auth/logout"
refresh_token = load_config("refreshToken")

if not refresh_token:
    print("Error: No refresh token found. Run A2.auth_login.py first.")
    sys.exit(1)

payload = {
    "refreshToken": refresh_token
}

response = send_and_print(
    url=url,
    method="POST",
    body=payload,
    output_file=f"{os.path.splitext(os.path.basename(__file__))[0]}.json"
)