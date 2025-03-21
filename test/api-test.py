import hashlib
import requests
import time

from rich.console import Console
from rich.json import JSON

BASE_URL = "http://localhost:5173"
console = Console()

def sleep_if_needed():
    time.sleep(2.0)

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
    console.print(f"[bold green]Create user '{username}' response:[/bold green]")
    console.print(JSON.from_data(response.json()))

    sleep_if_needed()
    return response.json()

def login_user(username, password):
    url = f"{BASE_URL}/api/user/login"
    payload = {
        "username": username,
        "password": hash_sha512(password)
    }
    response = requests.post(url, json=payload)

    console.print(f"[bold green]Login user '{username}' response:[/bold green]")
    console.print(JSON.from_data(response.json()))

    sleep_if_needed()
    return response.json()

def cash_in(session_token, security_code, amount):
    url = f"{BASE_URL}/api/cashin"
    headers = {
        "X-Session-Token": session_token,
        "X-Security-Code": security_code
    }
    payload = {"amount": str(amount)}
    response = requests.post(url, json=payload, headers=headers)

    console.print(f"[bold green]Cash in {amount} uro response:[/bold green]")
    console.print(JSON.from_data(response.json()))

    sleep_if_needed()
    return response.json()

def user_info(session_token, security_code):
    url = f"{BASE_URL}/api/user/info"
    headers = {
        "X-Session-Token": session_token,
        "X-Security-Code": security_code
    }
    response = requests.post(url, headers=headers)

    console.print("[bold green]User info response:[/bold green]")
    console.print(JSON.from_data(response.json()))

    sleep_if_needed()
    return response.json()

def payment_request(user_session, user_security, amount):
    url = f"{BASE_URL}/api/payment/request"
    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    payload = {"amount": str(amount)}

    response = requests.post(url, json=payload, headers=headers)

    console.print("[bold green]Payment request response:[/bold green]")
    console.print(JSON.from_data(response.json()))

    sleep_if_needed()
    return response.json()

def payment_send(user_session, user_security, transaction_id):
    url = f"{BASE_URL}/api/payment/send"
    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    payload = {"transaction_id": transaction_id}

    response = requests.post(url, json=payload, headers=headers)

    console.print("[bold green]Payment transaction response:[/bold green]")
    console.print(JSON.from_data(response.json()))

    sleep_if_needed()
    return response.json()

def withdraw(user_session, user_security, amount):
    url = f"{BASE_URL}/api/withdraw"
    headers = {
        "X-Session-Token": user_session,
        "X-Security-Code": user_security
    }
    payload = {"amount": str(amount)}

    response = requests.post(url, json=payload, headers=headers)

    console.print("[bold green]Withdraw response:[/bold green]")
    console.print(JSON.from_data(response.json()))

    sleep_if_needed()
    return response.json()

def user_logout(session_token):
    url = f"{BASE_URL}/api/user/logout"
    headers = {"X-Session-Token": session_token}

    response = requests.post(url, headers=headers)

    console.print("[bold green]Logout response:[/bold green]")
    console.print(JSON.from_data(response.json()))

    sleep_if_needed()
    return response.json()

def main():
    console.print("[bold blue]=== Creating Users ===[/bold blue]")

    alice_password = "hello01@@"
    bob_password = "hello02@@"

    create_user("Alice", "alice@example.com", alice_password)
    create_user("Bob", "bob@example.com", bob_password)

    console.print("\n[bold blue]=== Logging In Users ===[/bold blue]")
    alice_login = login_user("Alice", alice_password)
    bob_login = login_user("Bob", bob_password)

    alice_session = alice_login.get("session_token")
    bob_session = bob_login.get("session_token")

    alice_security = alice_login.get("security_code")
    bob_security = bob_login.get("security_code")

    console.print("\n[bold blue]=== Cash In Funds ===[/bold blue]")
    cash_in(alice_session, alice_security, 10000)
    cash_in(bob_session, bob_security, 5000)

    console.print("\n[bold blue]=== Fetching User Info ===[/bold blue]")
    user_info(alice_session, alice_security)
    user_info(bob_session, bob_security)

    console.print("\n[bold blue]=== Payment Request ===[/bold blue]")
    pr_response = payment_request(alice_session, alice_security, 1500)
    transaction_id = pr_response.get("transaction_id")

    if not transaction_id:
        console.print("[bold red]Payment request failed; no transaction_id returned.[/bold red]")
        return

    console.print("\n[bold blue]=== Payment Transaction ===[/bold blue]")
    payment_send(bob_session, bob_security, transaction_id)

    console.print("\n[bold blue]=== Withdraw Funds ===[/bold blue]")
    withdraw(alice_session, alice_security, 3000)

    console.print("\n[bold blue]=== Fetching User Info ===[/bold blue]")
    user_info(alice_session, alice_security)
    user_info(bob_session, bob_security)

    console.print("\n[bold blue]=== Logging Out Users ===[/bold blue]")
    user_logout(alice_session)
    user_logout(bob_session)

if __name__ == "__main__":
    main()
