package app

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username     string  `json:"username"`
	FullNames    string  `json:"full_names"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	Password     string  `json:"password"`
	Income       float64 `json:"income"`
	IncomeSource string  `json:"income_source"`
	IsTaxed      bool    `json:"is_taxed"`
	TaxRate      float64 `json:"tax_rate"`
	Currency     string  `json:"currency"`
	SpentRent    float64 `json:"spent_rent"`
	SpentFood    float64 `json:"spent_food"`
	SpentTrans   float64 `json:"spent_trans"`
	SpentUtils   float64 `json:"spent_utils"`
}

type BudgetApp struct {
	mu           sync.RWMutex
	users        map[string]User
	dbFile       string
	templateFile string
}

func NewBudgetApp(dbFile, templateFile string) *BudgetApp {
	app := &BudgetApp{
		users:        make(map[string]User),
		dbFile:       dbFile,
		templateFile: templateFile,
	}
	app.loadUsers()
	return app
}

func (app *BudgetApp) loadUsers() {
	app.mu.Lock()
	defer app.mu.Unlock()

	file, err := os.ReadFile(app.dbFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Printf("Error reading database: %v", err)
		return
	}

	if err := json.Unmarshal(file, &app.users); err != nil {
		log.Printf("Error unmarshaling database: %v", err)
	}
}

func (app *BudgetApp) saveUsers() error {
	data, err := json.MarshalIndent(app.users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(app.dbFile, data, 0o644)
}

func dashboardDisplayName(user User) string {
	if trimmed := strings.TrimSpace(user.Username); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(user.FullNames)
}

func currencyMeta(currencyCode string) (string, string) {
	code := strings.ToUpper(strings.TrimSpace(currencyCode))
	switch code {
	case "KSH", "KES", "KSHS":
		return "KSh", "Kenyan Shilling (KES)"
	case "USD", "US$":
		return "$", "US Dollar (USD)"
	case "EUR", "€":
		return "€", "Euro (EUR)"
	case "GBP", "£":
		return "£", "British Pound (GBP)"
	case "CAD":
		return "C$", "Canadian Dollar (CAD)"
	case "AUD":
		return "A$", "Australian Dollar (AUD)"
	case "JPY":
		return "¥", "Japanese Yen (JPY)"
	case "CHF":
		return "CHF", "Swiss Franc (CHF)"
	case "ZAR":
		return "R", "South African Rand (ZAR)"
	case "NGN":
		return "₦", "Nigerian Naira (NGN)"
	case "UGX":
		return "USh", "Ugandan Shilling (UGX)"
	case "TZS":
		return "TSh", "Tanzanian Shilling (TZS)"
	case "AED":
		return "د.إ", "UAE Dirham (AED)"
	case "INR":
		return "₹", "Indian Rupee (INR)"
	case "CNY":
		return "¥", "Chinese Yuan (CNY)"
	case "SEK":
		return "kr", "Swedish Krona (SEK)"
	case "NOK":
		return "kr", "Norwegian Krone (NOK)"
	case "DKK":
		return "kr", "Danish Krone (DKK)"
	case "PLN":
		return "zł", "Polish Złoty (PLN)"
	case "BRL":
		return "R$", "Brazilian Real (BRL)"
	case "MXN":
		return "$", "Mexican Peso (MXN)"
	case "SGD":
		return "S$", "Singapore Dollar (SGD)"
	case "PHP":
		return "₱", "Philippine Peso (PHP)"
	case "IDR":
		return "Rp", "Indonesian Rupiah (IDR)"
	case "MYR":
		return "RM", "Malaysian Ringgit (MYR)"
	case "THB":
		return "฿", "Thai Baht (THB)"
	case "VND":
		return "₫", "Vietnamese Dong (VND)"
	case "PKR":
		return "₨", "Pakistani Rupee (PKR)"
	case "EGP":
		return "E£", "Egyptian Pound (EGP)"
	case "MAD":
		return "د.م.", "Moroccan Dirham (MAD)"
	default:
		return currencyCode, currencyCode
	}
}

func greetingForCurrency(currencyCode string) string {
	code := strings.ToUpper(strings.TrimSpace(currencyCode))
	switch code {
	case "TZS":
		return "Karibu"
	case "UGX":
		return "Oli otya"
	case "KSH", "KES", "KSHS":
		return "Habari"
	case "USD", "CAD", "AUD", "NZD", "SGD", "HKD", "BND", "MXN", "BRL", "ARS", "GBP":
		return "Hello"
	case "EUR", "CHF", "SEK", "NOK", "DKK", "PLN", "CZK", "HUF":
		return "Hallo"
	case "NGN":
		return "Kedu"
	case "ZAR":
		return "Dumela"
	case "JPY":
		return "こんにちは"
	case "INR":
		return "नमस्ते"
	case "CNY":
		return "你好"
	case "AED", "SAR", "QAR", "BHD", "OMR", "KWD":
		return "مرحبا"
	case "PHP":
		return "Kamusta"
	case "IDR":
		return "Halo"
	case "THB":
		return "สวัสดี"
	case "VND":
		return "Xin chào"
	case "PKR":
		return "السلام علیکم"
	case "EGP":
		return "أهلًا"
	case "MAD":
		return "سلام"
	default:
		return "Welcome"
	}
}

type CurrencyAdvice struct {
	Intro      string
	SaccoTitle string
	SaccoBody  string
	MmfTitle   string
	MmfBody    string
	TBillTitle string
	TBillBody  string
}

func currencyAdviceForCurrency(currencyCode string) CurrencyAdvice {
	code := strings.ToUpper(strings.TrimSpace(currencyCode))
	switch code {
	case "KES", "KSH", "KSHS", "UGX", "TZS":
		return CurrencyAdvice{
			Intro:      "For East African currencies, a balanced plan usually combines a short-term emergency reserve with a longer-term savings habit. The examples below reflect common regional options for Kenya, Uganda, and Tanzania, including products that can send money to M-Pesa or a bank.",
			SaccoTitle: "SACCO savings",
			SaccoBody:  "Examples include Stima SACCO, Harambee SACCO, Mhasibu SACCO, and local savings groups. Many offer annual dividends or returns around 8%–15% depending on the society and its performance. They are great for disciplined saving, but you should check withdrawal rules and whether your shares can be redeemed quickly or need a notice period. Funds often reach your bank or M-Pesa in 1–3 business days after approval.",
			MmfTitle:   "Money market fund",
			MmfBody:    "Examples include Sanlam Investments, CIC Asset Management, Madison Asset Management, and other regional money market funds. They are useful for emergency reserves because they are more liquid than fixed deposits, with yields that can change with market conditions. Withdrawals often take 24–48 hours to your bank or M-Pesa, although some providers may take longer depending on their processing cycle.",
			TBillTitle: "Treasury bills or fixed deposits",
			TBillBody:  "Treasury bills, fixed deposits, and bank term accounts are useful when you want predictable returns and can leave money untouched for 91 days to 1 year or longer. They are usually lower risk than equities but less liquid than an MMF. Withdrawal or maturity payout often reaches your bank account within 1–3 business days after maturity, and some banks may take slightly longer.",
		}
	case "USD", "CAD", "AUD", "NZD", "SGD", "HKD", "BND", "MXN":
		return CurrencyAdvice{
			Intro:      "For major international currencies, a strong plan often combines a cash buffer, a market-linked growth option, and a long-term core holding. The examples below are common in the US, Canada, Australia, and similar markets.",
			SaccoTitle: "Credit union or cooperative account",
			SaccoBody:  "Credit unions and cooperatives often provide lower-fee savings products and community-based lending. They are useful when you want disciplined saving and a stable local institution, though returns can be lower than market funds.",
			MmfTitle:   "High-yield cash or money market account",
			MmfBody:    "High-yield savings accounts, money market funds, or cash management accounts are useful for emergency reserves. They usually offer liquidity with yields that can change monthly, and transfers to a bank may take 1–3 business days.",
			TBillTitle: "Treasury bills or ETFs",
			TBillBody:  "Treasury bills and broad index ETFs are common long-term choices. They tend to be more predictable than single stocks, but the right option depends on whether you want safety, income, or long-term growth.",
		}
	case "EUR", "GBP", "CHF", "SEK", "NOK", "DKK", "PLN", "CZK", "HUF":
		return CurrencyAdvice{
			Intro:      "For European and UK-style currencies, the best plan often balances liquidity with long-term growth. These examples reflect common options in Europe and the UK.",
			SaccoTitle: "Cooperative savings or local bank savings",
			SaccoBody:  "Cooperative banks and local savings institutions are common in Europe and the UK. They are useful when you want stability and a simple savings habit, though returns may be more conservative than growth funds.",
			MmfTitle:   "Money market fund or cash savings",
			MmfBody:    "European and UK money market products can be a strong emergency reserve. Withdrawals are usually reliable but can take 1–3 business days depending on your bank, broker, or provider.",
			TBillTitle: "Government bonds or index funds",
			TBillBody:  "Government bonds and low-fee index funds are common for medium-term planning. They can provide lower volatility than equity-heavy portfolios and are often suitable for people who want predictable growth.",
		}
	case "JPY":
		return CurrencyAdvice{
			Intro:      "For Japanese yen, a practical approach is often to keep a short-term cash buffer and then invest for longer-term stability. The examples below reflect typical Japanese savings and investment products.",
			SaccoTitle: "Credit union or postal savings",
			SaccoBody:  "Postal savings and regional credit unions are common for low-risk cash management. They are useful for keeping money safe while earning modest returns, though growth is usually slow and steady.",
			MmfTitle:   "Money market fund",
			MmfBody:    "Japanese money market funds can be used for liquidity and stability. Transfer and withdrawal times depend on the provider, but they are often slower than instant transfers and can take 1–2 business days.",
			TBillTitle: "Government bonds or diversified funds",
			TBillBody:  "Government bonds and diversified funds are strong choices for medium- to long-term stability. They are commonly used by people who want to protect capital while growing it steadily over time.",
		}
	case "INR":
		return CurrencyAdvice{
			Intro:      "For Indian rupees, a practical plan often combines a safety buffer with long-term investment products that are popular in India.",
			SaccoTitle: "Cooperative savings or local deposit account",
			SaccoBody:  "Cooperative and local bank savings accounts are a simple place to build discipline. They usually offer steadier returns than market-linked products and are good for short-term safety.",
			MmfTitle:   "Liquid mutual funds",
			MmfBody:    "Liquid funds are popular for emergency cash because they are easy to access and can be useful for short-term liquidity. Withdrawal times depend on the fund house and payment method, and some transfers may take a day or two.",
			TBillTitle: "Fixed deposits or index funds",
			TBillBody:  "Fixed deposits and broad index funds are common for medium-term planning. They are especially helpful when you want predictable returns and can leave the money invested for longer.",
		}
	case "CNY":
		return CurrencyAdvice{
			Intro:      "For Chinese yuan, a practical plan often balances stability with long-term wealth growth. Below are commonly used options in China and nearby financial markets.",
			SaccoTitle: "Local cooperative or bank savings",
			SaccoBody:  "Local bank savings and cooperative accounts are common for conservative savers. They are useful when you want a simple place to build cash reserves, though returns are typically modest.",
			MmfTitle:   "Money market fund",
			MmfBody:    "Money market funds are useful for relatively liquid cash reserves. They may offer better returns than ordinary saving accounts, but the exact yield and withdrawal timing depend on the fund provider and bank access.",
			TBillTitle: "Government bonds or diversified funds",
			TBillBody:  "Government bonds and diversified funds are usually better for longer-term planning. They can give you exposure to growth while keeping risk more controlled than buying single stocks.",
		}
	case "NGN":
		return CurrencyAdvice{
			Intro:      "For Nigerian naira, a strong plan usually balances liquidity with access to stable saving products that are common in Nigeria.",
			SaccoTitle: "Savings cooperative",
			SaccoBody:  "Savings cooperatives and mutual savings groups can be useful if you want a disciplined and community-based saving structure. Terms and returns vary, so review contribution rules and withdrawal conditions before you commit.",
			MmfTitle:   "Mutual fund or money market fund",
			MmfBody:    "Mutual funds and money market funds are widely used for emergency reserves and short-term goals. Processing times can be slower than instant transfers, and bank or app transfers may take 24–48 hours.",
			TBillTitle: "Treasury bills or fixed deposits",
			TBillBody:  "Treasury bills and fixed deposits are common for medium-term capital growth. They are helpful when you want stable returns and can leave the money untouched until maturity.",
		}
	case "ZAR":
		return CurrencyAdvice{
			Intro:      "For South African rand, a practical plan usually blends an emergency cash buffer with longer-term options that are common in South Africa.",
			SaccoTitle: "Savings cooperative",
			SaccoBody:  "Savings cooperatives and local fund institutions can be useful for disciplined saving. They can be a good fit if you want structure and community support, but returns and terms vary by provider.",
			MmfTitle:   "Money market or unit trust",
			MmfBody:    "Money market funds and unit trusts are often used for short-term liquidity and moderate returns. Withdrawal timing can depend on the provider and the payment rail, so check the terms before relying on fast access.",
			TBillTitle: "Government bonds or unit trusts",
			TBillBody:  "Government bonds and unit trusts are useful when you want lower volatility and can invest for longer. They can be a practical step up from purely cash-based savings.",
		}
	case "AED", "SAR", "QAR", "BHD", "OMR", "KWD":
		return CurrencyAdvice{
			Intro:      "For Gulf-region currencies, a strong plan often focuses on liquidity, predictable return products, and longer-term capital growth. The examples below reflect common regional options.",
			SaccoTitle: "Cooperative or savings account",
			SaccoBody:  "Cooperative and bank savings accounts remain a simple way to build discipline. They are useful when you want a stable base and predictable access to cash.",
			MmfTitle:   "Money market fund or cash management fund",
			MmfBody:    "Cash management and money market products are common for emergency reserves. They are generally lower risk than equities, but withdrawal speed and fees depend on the provider and the transfer method.",
			TBillTitle: "Government bonds or Islamic investment funds",
			TBillBody:  "Government bonds and sharia-compatible investment funds are common for medium- and long-term growth. They can give you stability while keeping exposure diversified.",
		}
	default:
		return CurrencyAdvice{
			Intro:      "A practical plan usually combines a reliable emergency reserve with a longer-term investment strategy. The best mix depends on your risk tolerance, timeline, and how quickly you may need to access cash.",
			SaccoTitle: "Cooperative or community savings",
			SaccoBody:  "Community-based savings structures can be useful when you want discipline, local support, and lower-risk participation. Review rates, fees, and withdrawal rules carefully.",
			MmfTitle:   "Money market or cash fund",
			MmfBody:    "Money market funds and cash funds are useful for short-term liquidity. They are often more flexible than fixed deposits, but transfer times depend on the provider and payment network.",
			TBillTitle: "Government bonds or diversified funds",
			TBillBody:  "Government bonds and diversified funds are suitable for longer-term planning. They are usually more stable than stock picking and can help protect capital while still allowing growth.",
		}
	}
}

func (app *BudgetApp) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles(app.templateFile)
		if err != nil {
			http.Error(w, "Failed to load index page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{"View": "signup"})
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		fullNames := r.FormValue("full_names")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		tmpl, err := template.ParseFiles(app.templateFile)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		if password != confirmPassword {
			tmpl.Execute(w, map[string]interface{}{"View": "signup", "Error": "Passwords do not match."})
			return
		}

		app.mu.Lock()
		defer app.mu.Unlock()

		if _, exists := app.users[username]; exists {
			tmpl.Execute(w, map[string]interface{}{"View": "signup", "Error": "Username is already taken."})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error processing password", http.StatusInternalServerError)
			return
		}

		app.users[username] = User{
			Username:     username,
			FullNames:    fullNames,
			Email:        email,
			Phone:        phone,
			Password:     string(hashedPassword),
			Income:       0.0,
			IncomeSource: "Salary",
			IsTaxed:      true,
			TaxRate:      0.0,
			Currency:     "KSh",
		}
		if err := app.saveUsers(); err != nil {
			log.Printf("failed to save users: %v", err)
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (app *BudgetApp) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles(app.templateFile)
		if err != nil {
			http.Error(w, "Failed to load login", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{"View": "login"})
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		tmpl, err := template.ParseFiles(app.templateFile)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		app.mu.RLock()
		user, exists := app.users[username]
		app.mu.RUnlock()

		if !exists {
			tmpl.Execute(w, map[string]interface{}{"View": "login", "Error": "Wrong username or password"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			tmpl.Execute(w, map[string]interface{}{"View": "login", "Error": "Wrong username or password"})
			return
		}

		http.Redirect(w, r, "/dashboard?user="+username, http.StatusSeeOther)
	}
}

func (app *BudgetApp) HandleUpdateRevenue(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("user")
	if r.Method == http.MethodPost && username != "" {
		incomeVal, _ := strconv.ParseFloat(r.FormValue("income"), 64)
		source := r.FormValue("income_source")
		isTaxed := r.FormValue("is_taxed") == "true"

		var taxVal float64
		if !isTaxed {
			taxVal, _ = strconv.ParseFloat(r.FormValue("tax_rate"), 64)
		} else {
			taxVal = 0.0
		}

		app.mu.Lock()
		if user, exists := app.users[username]; exists {
			user.Income = incomeVal
			user.IncomeSource = source
			user.IsTaxed = isTaxed
			user.TaxRate = taxVal
			app.users[username] = user
			if err := app.saveUsers(); err != nil {
				log.Printf("failed to save users: %v", err)
			}
		}
		app.mu.Unlock()
	}
	http.Redirect(w, r, "/dashboard?user="+username, http.StatusSeeOther)
}

func (app *BudgetApp) HandleDeductExpense(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("user")
	if r.Method == http.MethodPost && username != "" {
		target := r.FormValue("target")
		expenseVal, _ := strconv.ParseFloat(r.FormValue("expense"), 64)

		app.mu.Lock()
		if user, exists := app.users[username]; exists {
			switch target {
			case "rent":
				user.SpentRent = expenseVal
			case "food":
				user.SpentFood = expenseVal
			case "transport":
				user.SpentTrans = expenseVal
			case "utilities":
				user.SpentUtils = expenseVal
			}
			app.users[username] = user
			if err := app.saveUsers(); err != nil {
				log.Printf("failed to save users: %v", err)
			}
		}
		app.mu.Unlock()
	}
	http.Redirect(w, r, "/dashboard?user="+username, http.StatusSeeOther)
}

func (app *BudgetApp) HandleProfileUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		fullNames := r.FormValue("full_names")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		currency := r.FormValue("currency")
		customCurrency := r.FormValue("custom_currency")

		finalCurrency := currency
		if currency == "custom" && customCurrency != "" {
			finalCurrency = customCurrency
		}

		app.mu.Lock()
		if user, exists := app.users[username]; exists {
			user.FullNames = fullNames
			user.Email = email
			user.Phone = phone
			user.Currency = finalCurrency
			app.users[username] = user
			if err := app.saveUsers(); err != nil {
				log.Printf("failed to save users: %v", err)
			}
		}
		app.mu.Unlock()

		http.Redirect(w, r, "/dashboard?user="+username, http.StatusSeeOther)
	}
}

func (app *BudgetApp) HandleReset(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("user")
	if r.Method == http.MethodPost && username != "" {
		app.mu.Lock()
		if user, exists := app.users[username]; exists {
			user.Income = 0.0
			user.TaxRate = 0.0
			user.IsTaxed = true
			user.SpentRent = 0.0
			user.SpentFood = 0.0
			user.SpentTrans = 0.0
			user.SpentUtils = 0.0
			app.users[username] = user
			if err := app.saveUsers(); err != nil {
				log.Printf("failed to save users: %v", err)
			}
		}
		app.mu.Unlock()
	}
	http.Redirect(w, r, "/dashboard?user="+username, http.StatusSeeOther)
}

func (app *BudgetApp) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("user")
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.mu.RLock()
	user, exists := app.users[username]
	app.mu.RUnlock()

	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var netIncome float64
	if !user.IsTaxed && user.TaxRate > 0 {
		taxFactor := (100.0 - user.TaxRate) / 100.0
		netIncome = user.Income * taxFactor
	} else {
		netIncome = user.Income
	}

	needs := netIncome * 0.50
	baseSavings := netIncome * 0.20

	rentAlloc := needs * 0.40
	foodAlloc := needs * 0.30
	transAlloc := needs * 0.15
	utilsAlloc := needs * 0.15

	remRent := rentAlloc - user.SpentRent
	remFood := foodAlloc - user.SpentFood
	remTrans := transAlloc - user.SpentTrans
	remUtils := utilsAlloc - user.SpentUtils

	overspentRent := remRent < 0
	overspentFood := remFood < 0
	overspentTrans := remTrans < 0
	overspentUtils := remUtils < 0

	var rentExcess, foodExcess, transExcess, utilsExcess float64
	var totalDeficit float64

	if overspentRent {
		rentExcess = -remRent
		totalDeficit += rentExcess
	}
	if overspentFood {
		foodExcess = -remFood
		totalDeficit += foodExcess
	}
	if overspentTrans {
		transExcess = -remTrans
		totalDeficit += transExcess
	}
	if overspentUtils {
		utilsExcess = -remUtils
		totalDeficit += utilsExcess
	}

	isOverspent := overspentRent || overspentFood || overspentTrans || overspentUtils

	baseWants := netIncome * 0.30
	finalWants := baseWants - totalDeficit
	var deficitUnpaid float64
	if finalWants < 0 {
		deficitUnpaid = -finalWants
		finalWants = 0.0
	}

	var leftoverRent float64
	if !overspentRent && user.SpentRent < rentAlloc && user.SpentRent > 0 {
		leftoverRent = rentAlloc - user.SpentRent
	}

	savingsBalance := baseSavings + leftoverRent
	totalInvestable := savingsBalance

	saccoAlloc := totalInvestable * 0.40
	mmfAlloc := totalInvestable * 0.30
	tBillAlloc := totalInvestable * 0.30

	greeting := greetingForCurrency(user.Currency)

	currencySymbol, currencyLabel := currencyMeta(user.Currency)
	currencyAdvice := currencyAdviceForCurrency(user.Currency)
	data := map[string]interface{}{
		"View":              "dashboard",
		"Username":          user.Username,
		"FullNames":         user.FullNames,
		"DisplayName":       dashboardDisplayName(user),
		"Email":             user.Email,
		"Phone":             user.Phone,
		"RawIncome":         user.Income,
		"IncomeSource":      user.IncomeSource,
		"IsTaxed":           user.IsTaxed,
		"TaxRate":           user.TaxRate,
		"NetIncome":         netIncome,
		"Currency":          user.Currency,
		"CurrencySymbol":    currencySymbol,
		"CurrencyLabel":     currencyLabel,
		"Greeting":          greeting,
		"AdvisorIntro":      currencyAdvice.Intro,
		"AdvisorSaccoTitle": currencyAdvice.SaccoTitle,
		"AdvisorSaccoBody":  currencyAdvice.SaccoBody,
		"AdvisorMmfTitle":   currencyAdvice.MmfTitle,
		"AdvisorMmfBody":    currencyAdvice.MmfBody,
		"AdvisorTBillTitle": currencyAdvice.TBillTitle,
		"AdvisorTBillBody":  currencyAdvice.TBillBody,
		"Needs":             needs,
		"Wants":             finalWants,
		"BaseWants":         baseWants,
		"BaseSavings":       baseSavings,
		"SavingsBalance":    savingsBalance,
		"RentAlloc":         rentAlloc,
		"FoodAlloc":         foodAlloc,
		"TransAlloc":        transAlloc,
		"UtilsAlloc":        utilsAlloc,
		"SpentRent":         user.SpentRent,
		"SpentFood":         user.SpentFood,
		"SpentTrans":        user.SpentTrans,
		"SpentUtils":        user.SpentUtils,
		"RemRent":           remRent,
		"RemFood":           remFood,
		"RemTrans":          remTrans,
		"RemUtils":          remUtils,
		"LeftoverRent":      leftoverRent,
		"TotalInvestable":   totalInvestable,
		"SaccoAlloc":        saccoAlloc,
		"MmfAlloc":          mmfAlloc,
		"TBillAlloc":        tBillAlloc,
		"IsOverspent":       isOverspent,
		"OverspentRent":     overspentRent,
		"OverspentFood":     overspentFood,
		"OverspentTrans":    overspentTrans,
		"OverspentUtils":    overspentUtils,
		"RentExcess":        rentExcess,
		"FoodExcess":        foodExcess,
		"TransExcess":       transExcess,
		"UtilsExcess":       utilsExcess,
		"TotalDeficit":      totalDeficit,
		"DeficitUnpaid":     deficitUnpaid,
	}

	tmpl, err := template.ParseFiles(app.templateFile)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func ResolveProjectRoot(start string) string {
	if start == "" {
		start = "."
	}
	if _, err := os.Stat(filepath.Join(start, "go.mod")); err == nil {
		return start
	}
	if _, err := os.Stat(filepath.Join(start, "web", "templates", "index.html")); err == nil {
		return start
	}
	parent := filepath.Dir(start)
	if parent == start {
		return "."
	}
	return ResolveProjectRoot(parent)
}

func ResolveDataPath(root string) (string, string) {
	if root == "" {
		root = ResolveProjectRoot(".")
	}
	dbFile := filepath.Join(root, "internal", "data", "users.json")
	templateFile := filepath.Join(root, "web", "templates", "index.html")
	return dbFile, templateFile
}

func NewAppFromProjectRoot(root string) (*BudgetApp, error) {
	if root == "" {
		root = ResolveProjectRoot(".")
	}
	dbFile, templateFile := ResolveDataPath(root)
	if _, err := os.Stat(templateFile); err != nil {
		return nil, fmt.Errorf("template file not found: %w", err)
	}
	return NewBudgetApp(dbFile, templateFile), nil
}
