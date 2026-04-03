package authentication

import (
    "database/sql"
    "fmt"
    "log"
    "unicard-go/internal/pkg/account"
)

// CreateAccount handles all the business rules, ID generation, and DB transactions
func (h *Handler) CreateAccount(req SignupRequest) error {
    
    // 1. Check Card Status
    var cardStatus string
    err := h.DB.QueryRow("SELECT status FROM cards WHERE card_number = ?", req.CardNumber).Scan(&cardStatus)
    if err == sql.ErrNoRows {
        return fmt.Errorf("Card number not found. Please check your card")
    } else if err != nil {
        log.Printf("System error checking card status: %v", err)
        return fmt.Errorf("System error checking card status")
    }
    if cardStatus != "Inactive" {
        return fmt.Errorf("Card is currently '%s'. Please contact support", cardStatus)
    }

    // 2. Check Email
    exists, err := account.IsEmailExist(h.DB, req.Email)
    if err != nil {
        return fmt.Errorf("System error checking email")
    }
    if exists {
        return fmt.Errorf("Email already registered. Please use a different email")
    }

    // 3. Check Phone (Using your new repository function!)
    exists, err = h.isPhoneExist(req.ContactNumber)
    if err != nil {
        return fmt.Errorf("System error checking phone number")
    }
    if exists {
        return fmt.Errorf("Phone number already registered")
    }

    // 4. Hash Password
    hashedPassword, err := account.HashPassword(req.Password)
    if err != nil {
        return fmt.Errorf("System error processing password")
    }

    // 5. Generate IDs (Using your new utils functions!)
    generatedUsername, err := h.GenerateUniqueUsername()
    if err != nil { return fmt.Errorf("System error generating username") }
    
    generateUserId, err := h.GenerateUserID()
    if err != nil { return fmt.Errorf("System error generating UserID") }

    generateCardID, err := h.GenerateCardID()
    if err != nil { return fmt.Errorf("System error generating CardID") }

    createdAt, err := CurrentTimestamp()
    if err != nil { return fmt.Errorf("System error getting timestamp") }

    balance, err := h.GetInitialBalance(req.CardNumber)
    if err != nil { return fmt.Errorf("Invalid Card Number") }

    // 6. Build User struct
    user := User{
        UserID:     fmt.Sprintf("%d", generateUserId),
        CardID:     generateCardID,
        Usertype:   "Regular",
        Username:   generatedUsername,
        Fullname:   req.FirstName + " " + req.LastName,
        CardNumber: req.CardNumber,
        Password:   hashedPassword,
        Email:      req.Email,
        Phone:      req.ContactNumber,
        CreatedAt:  createdAt,
        Balance:    balance,
    }

    // 7. Begin transaction: insert user + activate card atomically
    tx, err := h.DB.Begin()
    if err != nil {
        return fmt.Errorf("System error starting transaction")
    }
    defer tx.Rollback()

    insertQuery := `INSERT INTO users (user_id, username, full_name, email, phone, password_hash, card_id, card_number, user_type, balance, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    _, err = tx.Exec(insertQuery, user.UserID, user.Username, user.Fullname, user.Email, user.Phone, user.Password, user.CardID, user.CardNumber, user.Usertype, user.Balance, user.CreatedAt)
    if err != nil {
        log.Printf("Error inserting user: %v", err)
        return fmt.Errorf("System error creating account. Please try again")
    }

    _, err = tx.Exec("UPDATE cards SET status = 'Active' WHERE card_number = ?", user.CardNumber)
    if err != nil {
        log.Printf("Error activating card: %v", err)
        return fmt.Errorf("System error activating card")
    }

    if err = tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
        return fmt.Errorf("System error finalizing account creation")
    }

    log.Printf("Account successfully created! UserID: %s", user.UserID)
    return nil // Success! No errors!
}