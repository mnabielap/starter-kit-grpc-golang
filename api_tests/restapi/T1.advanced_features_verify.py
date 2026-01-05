import sys
import os
import time
import json
from datetime import datetime
from urllib.parse import quote

# Add current directory to path to import utils
sys.path.append(os.path.abspath(os.path.dirname(__file__)))
from utils import send_and_print, BASE_URL, load_config

# --- CONFIGURATION ---
TIMESTAMP = int(time.time())

# SET THIS BASED ON YOUR DATABASE TYPE
# True  = Database uses ENUM (Order: User -> Admin)
# False = Database uses String/VarChar (Order: Admin -> User [Alphabetical])
IS_ROLE_ENUM = False 

TEST_USERS = [
    # Alice (Admin)
    {"name": f"AutoTest Alice {TIMESTAMP}", "email": f"alice.{TIMESTAMP}@test.com", "role": "admin", "password": "password123"},
    # Bob (User)
    {"name": f"AutoTest Bob {TIMESTAMP}", "email": f"bob.{TIMESTAMP}@test.com", "role": "user", "password": "password123"},
    # Charlie (User)
    {"name": f"AutoTest Charlie {TIMESTAMP}", "email": f"charlie.{TIMESTAMP}@test.com", "role": "user", "password": "password123"},
]
CREATED_USERS = [] 

# --- COLORS & HELPERS ---
class Colors:
    HEADER = '\033[95m'
    OKGREEN = '\033[92m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_header(msg):
    print(f"\n{Colors.HEADER}{Colors.BOLD}=== {msg} ==={Colors.ENDC}")

def print_pass(msg):
    print(f"{Colors.OKGREEN}[PASS] {msg}{Colors.ENDC}")

def print_fail(msg):
    print(f"{Colors.FAIL}[FAIL] {msg}{Colors.ENDC}")

def get_token():
    token = load_config("accessToken")
    if not token:
        print_fail("No access token found. Please run A2.auth_login.py first.")
        sys.exit(1)
    return token

def snake_to_camel(snake_str):
    """Converts snake_case to camelCase (e.g. created_at -> createdAt)"""
    components = snake_str.split('_')
    return components[0] + ''.join(x.title() for x in components[1:])

def get_json_value(item, key):
    """
    Safely retrieves value from dict supporting both snake_case (Proto default) 
    and camelCase (Gateway JSON default).
    """
    # 1. Try exact match
    if key in item:
        return item[key]
    
    # 2. Try camelCase conversion (Gateway output)
    camel_key = snake_to_camel(key)
    if camel_key in item:
        return item[camel_key]
    
    # 3. Try PascalCase (Unlikely but good for safety)
    pascal_key = camel_key[0].upper() + camel_key[1:]
    if pascal_key in item:
        return item[pascal_key]

    raise KeyError(f"Key '{key}' or '{camel_key}' not found in item: {list(item.keys())}")

# --- 1. SEEDING ---
def create_seed_users(token):
    print_header("1. SEEDING DATA")
    url = f"{BASE_URL}/users"
    headers = {"Authorization": f"Bearer {token}"}
    
    for user_data in TEST_USERS:
        resp = send_and_print(url, headers, method="POST", body=user_data, output_file="temp_seed.json")
        if resp.status_code == 201:
            full_user = resp.json()
            # Normalize ID to string
            full_user['id'] = str(full_user['id']) 
            CREATED_USERS.append(full_user)
            print(f"Created: {full_user['name']} (ID: {full_user['id']})")
            time.sleep(1) # Ensure timestamps are different for sorting tests
        else:
            print_fail(f"Failed to create user {user_data['name']}")
            cleanup_users(token)
            sys.exit(1)

def cleanup_users(token):
    print_header("CLEANUP")
    headers = {"Authorization": f"Bearer {token}"}
    for user in CREATED_USERS:
        url = f"{BASE_URL}/users/{user['id']}"
        send_and_print(url, headers, method="DELETE", output_file="temp_cleanup.json")
        print(f"Deleted user ID: {user['id']}")

# --- 2. SEARCH TESTS ---
def test_search_scopes(token):
    print_header("2. SEARCH SCOPES")
    headers = {"Authorization": f"Bearer {token}"}
    search_base = quote(str(TIMESTAMP))

    # 2.1 Scope: ALL
    print(f">> Case A: Scope 'all'")
    url = f"{BASE_URL}/users?search={search_base}&scope=all&limit=100"
    resp = send_and_print(url, headers, output_file="test_search_all.json")
    
    results = resp.json().get('results', [])
    found_ids = [str(u['id']) for u in results]
    expected_ids = [str(u['id']) for u in CREATED_USERS]
    
    if all(uid in found_ids for uid in expected_ids):
        print_pass("Found all seeded users.")
    else:
        print_fail(f"Scope 'all' failed.")

    # 2.2 Scope: NAME
    target = CREATED_USERS[0] 
    search_name = quote(target['name'])
    print(f">> Case B: Scope 'name'")
    url = f"{BASE_URL}/users?search={search_name}&scope=name"
    resp = send_and_print(url, headers, output_file="test_search_name.json")
    results = resp.json().get('results', [])
    
    if len(results) == 1 and str(results[0]['id']) == str(target['id']):
        print_pass("Found user by Name.")
    else:
        print_fail("Failed to find user by Name.")

    # 2.3 Scope: EMAIL
    target = CREATED_USERS[1] 
    search_email = quote(target['email'])
    print(f">> Case C: Scope 'email'")
    url = f"{BASE_URL}/users?search={search_email}&scope=email"
    resp = send_and_print(url, headers, output_file="test_search_email.json")
    results = resp.json().get('results', [])
    
    if len(results) == 1 and str(results[0]['id']) == str(target['id']):
        print_pass("Found user by Email.")
    else:
        print_fail("Failed to find user by Email.")

    # 2.4 Scope: ID
    target = CREATED_USERS[2] 
    print(f">> Case D: Scope 'id'")
    url = f"{BASE_URL}/users?search={target['id']}&scope=id"
    resp = send_and_print(url, headers, output_file="test_search_id.json")
    results = resp.json().get('results', [])
    
    if len(results) == 1 and str(results[0]['id']) == str(target['id']):
        print_pass("Found user by ID.")
    else:
        print_fail(f"Failed to find user by ID.")

# --- 3. FILTER TESTS ---
def test_filters(token):
    print_header("3. ROLE FILTERING")
    headers = {"Authorization": f"Bearer {token}"}
    search_base = quote(str(TIMESTAMP))
    
    print(">> Case A: Role 'admin'")
    url = f"{BASE_URL}/users?role=admin&search={search_base}&scope=all"
    resp = send_and_print(url, headers, output_file="test_filter_admin.json")
    results = resp.json().get('results', [])
    
    has_alice = any(str(u['id']) == str(CREATED_USERS[0]['id']) for u in results)
    has_bob = any(str(u['id']) == str(CREATED_USERS[1]['id']) for u in results)
    
    if has_alice and not has_bob:
        print_pass("Filter 'admin' passed.")
    else:
        print_fail("Filter 'admin' failed.")

# --- 4. SORTING TESTS ---
def test_sorting(token):
    print_header(f"4. SORTING")
    headers = {"Authorization": f"Bearer {token}"}
    search_base = quote(str(TIMESTAMP))
    
    def check_sort_order(field, order):
        print(f">> Testing Sort: {field} ({order.upper()})")
        
        url = f"{BASE_URL}/users?sort={field}:{order}&search={search_base}&scope=all"
        
        resp = send_and_print(url, headers, output_file=f"test_sort_{field}.json")
        results = resp.json().get('results', [])
        
        if len(results) < 2:
            print_fail("Not enough results to verify sort.")
            return

        try:
            api_values = [get_json_value(r, field) for r in results]
        except KeyError as e:
            print_fail(f"Key Error: {e}")
            return

        # Sorting Strategy
        if field == 'id':
            # UUID sorting is just string sorting
            expected_values = sorted(api_values, key=str, reverse=(order == 'desc'))
        elif field == 'role' and IS_ROLE_ENUM:
            role_weight = {"user": 1, "admin": 2}
            expected_values = sorted(api_values, key=lambda x: role_weight.get(x, 99), reverse=(order == 'desc'))
        else:
            expected_values = sorted(api_values, reverse=(order == 'desc'))

        if api_values == expected_values:
            print_pass(f"Sort {field} {order} is correct.")
        else:
            print_fail(f"Sort {field} {order} FAILED.")
            print(f"   API Got:  {api_values}")
            print(f"   Expected: {expected_values}")

    # Run Sort Tests
    check_sort_order('id', 'asc')
    check_sort_order('id', 'desc')
    check_sort_order('name', 'asc')
    check_sort_order('name', 'desc')
    check_sort_order('email', 'asc')
    check_sort_order('email', 'desc')
    check_sort_order('created_at', 'asc') # Will look for 'createdAt' in JSON
    check_sort_order('created_at', 'desc')
    check_sort_order('role', 'asc')
    check_sort_order('role', 'desc')

# --- 5. PAGINATION TESTS ---
def test_pagination(token):
    print_header("5. PAGINATION")
    headers = {"Authorization": f"Bearer {token}"}
    search_base = quote(str(TIMESTAMP))

    base_url = f"{BASE_URL}/users?sort=created_at:asc&search={search_base}&scope=all"

    # 5.1 Page 1 (Alice)
    print(">> Case A: Page 1, Limit 1")
    url = f"{base_url}&limit=1&page=1"
    resp = send_and_print(url, headers, output_file="test_page_1.json")
    results = resp.json().get('results', [])
    
    if len(results) == 1 and str(results[0]['id']) == str(CREATED_USERS[0]['id']):
        print_pass("Page 1 correctly returned Alice.")
    else:
        print_fail(f"Page 1 failed.")

    # 5.2 Page 2 (Bob)
    print(">> Case B: Page 2, Limit 1")
    url = f"{base_url}&limit=1&page=2"
    resp = send_and_print(url, headers, output_file="test_page_2.json")
    results = resp.json().get('results', [])
    
    if len(results) == 1 and str(results[0]['id']) == str(CREATED_USERS[1]['id']):
        print_pass("Page 2 correctly returned Bob.")
    else:
        print_fail(f"Page 2 failed.")

    # 5.3 Page 3 (Charlie)
    print(">> Case C: Page 3, Limit 1")
    url = f"{base_url}&limit=1&page=3"
    resp = send_and_print(url, headers, output_file="test_page_3.json")
    results = resp.json().get('results', [])

    if len(results) == 1 and str(results[0]['id']) == str(CREATED_USERS[2]['id']):
        print_pass("Page 3 correctly returned Charlie.")
    else:
        print_fail("Page 3 failed.")


# --- MAIN EXECUTION ---
if __name__ == "__main__":
    try:
        access_token = get_token()
        create_seed_users(access_token)
        test_search_scopes(access_token)
        test_filters(access_token)
        test_sorting(access_token)
        test_pagination(access_token)
    except Exception as e:
        print(f"\n{Colors.FAIL}[FAIL] CRITICAL ERROR: {e}{Colors.ENDC}")
        import traceback
        traceback.print_exc()
    finally:
        try:
            cleanup_users(access_token)
        except:
            pass