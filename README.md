# V2X Post-Quantum Onboarding Simulation

> A Proof of Concept (PoC) simulating the secure onboarding process of autonomous vehicles (OBU) into road infrastructure (RSU) within an ITS/V2X environment. This project utilizes Post-Quantum Cryptography (PQC) mechanisms, specifically Key Encapsulation Mechanisms (KEM), to secure key exchange against potential quantum computer attacks.

The architecture consists of a high-performance backend written in **Go** (handling cryptographic logic and network simulation) and an interactive analytical dashboard built in **React**, which visualizes the message exchange process via WebSockets.

## 🛠 Technologies & Architecture

* **Backend:** Go (Golang), `gorilla/websocket`
* **Cryptography (PQC):** Open Quantum Safe Project (`liboqs-go`)
* **Frontend:** React, TypeScript, WebSockets
* **Communication:** Asynchronous event flow OBU <-> RSU

## ⚙️ Prerequisites

Before running the project locally, ensure you have the following tools installed:

1. **Go** (v1.20 or newer)
2. **Node.js & npm** (for the React frontend)
3. **`make` tool** (for build automation)
4. **liboqs (CRITICAL):** The Go backend requires the `liboqs` C library installed on your system. 
   * See installation instructions: [open-quantum-safe/liboqs](https://github.com/open-quantum-safe/liboqs).

---

## 🚀 Development Commands

The project uses a `Makefile` to automate the workflow for both Go and React environments:

* **Build and Test:** Runs the complete build process for the backend, installs frontend dependencies, and executes unit tests (including KEM operation tests).
  ```bash
  make all

* **Build Application:** Compiles Go binaries and builds static React files.
  ```bash
    make build

* **Run Application:** Starts the backend server (listening for WebSockets) and the React dev server (localhost:5173).
  ```bash
  make run

* **Live Reload (Watch):** Runs the project in development mode, automatically reloading on file changes.
  ```bash
  make watch
  
* **Run Tests:** Executes tests verifying encryption, decapsulation logic, and WebSocket communication.
  ```bash
  make test
  
* **Clean Build Artifacts:** Removes binaries and build folders to ensure a fresh environment.
  ```bash
  make clean
