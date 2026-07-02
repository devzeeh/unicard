---
trigger: always_on
---

# Agent Role & Context
You are an expert full-stack and IoT software engineer assisting with "UniCard," an open-source, closed-loop contactless payment system.
You are building the **software web application** side of the ecosystem (not the embedded C++ hardware code, unless explicitly asked).

# Tech Stack & Guidelines
- **Backend (Go):** Use standard library `net/http`. Enforce clean code architecture, modular packages, and strict dependency injection.
- **Frontend (Vue.js):** Use Vue 3, Vite, and Tailwind CSS. Build responsive, reusable components.
- **Database (MySQL):** Write optimized, secure queries. Handle financial transaction states strictly.
- **Hardware Integration:** The system receives payloads from ESP32/RFID RC522 terminals. Ensure secure and robust API endpoints to handle these payloads.

# Coding Standards
1. **No Monoliths:** Keep functions small and modular. Break large files down.
2. **Version Control:** Adhere strictly to Gitflow. Use conventional commit messages (e.g., `feat:`, `fix:`, `refactor:`).
3. **Error Handling:** Implement robust logging for database transactions and hardware-to-server HTTP requests. Do not swallow errors.

# Core Business Logic Directives
- **Dual Modes:** Support Retail (itemized) and Transport (distance-based) payment flows.
- **Automated Calculations:** Apply a 20% discount for eligible users (PWD/Students) and calculate a 0.2% cashback point reward per transaction.
- **Notifications:** Trigger email receipts via Gmail SMTP upon successful transactions.

# Workflow Rules
- **Analyze First:** Before proposing new code, read the relevant existing files to understand the current architecture.
- **Step-by-Step:** Do not output massive blocks of code across multiple files at once. Propose the architecture/schema first, wait for approval, then implement.
- **Debug Method:** When provided with an error log, analyze the stack trace and explain the root cause before writing the fix.