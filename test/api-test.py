import hashlib
import requests
import time

BASE_URL = "http://localhost:5173"

def sleep_if_needed():
    time.sleep(2.1)

def hash_sha512(input_str: str):
    return hashlib.sha512(input_str.encode()).hexdigest()

def create_user(username, email, password):
    url = f"{BASE_URL}/api/user/create"

    payload = {
        "username": username,
        "email": email,
        "password": hash_sha512(password)
    }
    response = requests.post(url, json=payload)

    print(f"Create user '{username}' response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def login_user(username, password):
    url = f"{BASE_URL}/api/user/login"

    payload = {
        "username": username,
        "password": hash_sha512(password)
    }
    response = requests.post(url, json=payload)

    print(f"Login user '{username}' response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def cash_in(session_token, security_code, amount):
    url = f"{BASE_URL}/api/cashin"

    headers = {
        "X-Session-Token": session_token,
        "X-Security-Code": security_code
    }
    payload = {"amount": amount}
    response = requests.post(url, json=payload, headers=headers)

    print(f"Cash in {amount} uro response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def user_info(session_token, security_code):
    url = f"{BASE_URL}/api/user/info"

    headers = {
        "X-Session-Token": session_token,
        "X-Security-Code": security_code
    }
    response = requests.post(url, headers=headers)

    print("User info response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def loan_request(
    debtor_session,
    debtor_security,
    creditor_identifier,
    amount,
    loan_type,
    timespan_days,
    payment_type
):
    url = f"{BASE_URL}/api/loan/request"

    headers = {
        "X-Session-Token": debtor_session,
        "X-Security-Code": debtor_security
    }
    payload = {
        "creditor_identifier": creditor_identifier,
        "amount": amount,
        "loan_type": loan_type,
        "timespan_days": timespan_days,
        "payment_type": payment_type
    }
    response = requests.post(url, json=payload, headers=headers)

    print("Loan request response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def loan_accept(creditor_session, creditor_security, loan_id):
    url = f"{BASE_URL}/api/loan/accept"

    headers = {
        "X-Session-Token": creditor_session,
        "X-Security-Code": creditor_security
    }
    payload = {"loan_id": loan_id}
    response = requests.post(url, json=payload, headers=headers)

    print("Loan accept response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def loan_reject(creditor_session, creditor_security, loan_id, rejection_code, message):
    url = f"{BASE_URL}/api/loan/reject"

    headers = {
        "X-Session-Token": creditor_session,
        "X-Security-Code": creditor_security
    }
    payload = {
        "loan_id": loan_id,
        "rejection_code": rejection_code,
        "message": message
    }
    response = requests.post(url, json=payload, headers=headers)

    print("Loan reject response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def payment_transaction(payer_session, payer_security, recipient_identifier, amount):
    url = f"{BASE_URL}/api/payment/transaction"

    headers = {
        "X-Session-Token": payer_session,
        "X-Security-Code": payer_security
    }
    payload = {
        "recipient_identifier": recipient_identifier,
        "amount": amount
    }
    response = requests.post(url, json=payload, headers=headers)

    print("Payment transaction response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def payment_request(user_session, user_security, amount):
    url = f"{BASE_URL}/api/payment/request"

    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    payload = {"amount": amount}
    response = requests.post(url, json=payload, headers=headers)

    print("Payment request response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def refund_request(user_session, user_security, loan_id):
    url = f"{BASE_URL}/api/refund/request"

    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    payload = {"loan_id": loan_id}
    response = requests.post(url, json=payload, headers=headers)

    print("Refund request response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def refund_reject(user_session, user_security, refund_id, message):
    url = f"{BASE_URL}/api/refund/reject"

    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    payload = {
        "refund_id": refund_id,
        "message": message
    }
    response = requests.post(url, json=payload, headers=headers)

    print("Refund reject response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def refund_process(user_session, user_security, refund_id):
    url = f"{BASE_URL}/api/refund/process"

    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    payload = {"refund_id": refund_id}
    response = requests.post(url, json=payload, headers=headers)

    print("Refund process response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def withdraw(user_session, user_security, amount):
    url = f"{BASE_URL}/api/withdraw"

    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    payload = {"amount": amount}
    response = requests.post(url, json=payload, headers=headers)

    print("Withdraw response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def user_logout(session_token):
    url = f"{BASE_URL}/api/user/logout"

    headers = {"X-Session-Token": session_token}
    response = requests.post(url, headers=headers)

    print("Logout response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def fetch_notifications(user_session, user_security):
    url = f"{BASE_URL}/api/user/notifications"

    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    response = requests.post(url, headers=headers)

    print("Notifications response:")
    print(response.json())

    sleep_if_needed()
    return response.json()

def main():
    print("=== Creating Users ===")

    alice_password = "hello01@@"
    bob_password = "hello02@@"

    alice = create_user("Alice", "alice@example.com", alice_password)
    bob = create_user("Bob", "bob@example.com", bob_password)

    alice_identifier = alice.get("identifier")
    bob_identifier = bob.get("identifier")

    print("\n=== Logging In Users ===")

    alice_login = login_user("Alice", alice_password)
    bob_login = login_user("Bob", bob_password)

    alice_session = alice_login.get("session_token")
    bob_session = bob_login.get("session_token")

    alice_security = alice_login.get("security_code")
    bob_security = bob_login.get("security_code")

    print("Alice Security Code: ", alice_security)
    print("Bob Security Code: ", bob_security)

    print("\n=== Cash In Funds ===")
    cash_in(alice_session, alice_security, 10000)
    cash_in(bob_session, bob_security, 5000)

    print("\n=== Fetching User Info ===")
    user_info(alice_session, alice_security)
    user_info(bob_session, bob_security)

    print("\n=== Loan Request & Acceptance ===")
    loan_req = loan_request(
        debtor_session=alice_session,
        debtor_security=alice_security,
        creditor_identifier=bob_identifier,
        amount=2000,
        loan_type="uro",
        timespan_days=30,
        payment_type="standard"
    )

    loan_id = loan_req.get("loan_id")
    if loan_id:
        loan_accept(bob_session, bob_security, loan_id)
    else:
        print("Loan request failed. Skipping acceptance.")

    print("\n=== Payment Transaction ===")
    payment_transaction(alice_session, alice_security, bob_identifier, 1500)

    print("\n=== Payment Request ===")
    payment_request(alice_session, alice_security, 800)

    print("\n=== Refund Request, Reject & Process ===")
    refund_req = refund_request(alice_session, alice_security, loan_id)
    refund_id = refund_req.get("refund_id")

    if refund_id:
        refund_reject(bob_session, bob_security, refund_id, "Not valid refund")
        refund_process(bob_session, bob_security, refund_id)
    else:
        print("Refund request failed. Skipping refund tests.")

    print("\n=== Withdraw Funds ===")
    withdraw(alice_session, alice_security, 3000)

    print("\n=== Fetching Notifications ===")
    fetch_notifications(bob_session, bob_security)

    print("\n=== Logging Out Users ===")
    user_logout(alice_session)
    user_logout(bob_session)

if __name__ == "__main__":
    main()
