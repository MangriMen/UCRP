package seed

import "fmt"

const guestSeedSalt = 1000003

type identifierSpec struct {
	Type   string
	Value  string
	Source string
}

type eventSpec struct {
	GuestKey   string
	Tenant     string
	Project    string
	Type       string
	Source     string
	DaysAgo    int
	Hour       int
	Minute     int
	Amount     int64
	HasAmount  bool
	ExternalID string
	Reverses   string
	Keys       []string
}

type guestSpec struct {
	Key            string
	Name           string
	Status         string
	MergedInto     string
	CreatedDaysAgo int
	Identifiers    []identifierSpec
	Events         []eventSpec
}

func backgroundGuestKey(i int) string {
	return fmt.Sprintf("bg-%03d", i)
}

var firstNames = []string{
	"Александр", "Мария", "Дмитрий", "Елена", "Сергей", "Ольга", "Андрей", "Наталья",
	"Иван", "Татьяна", "Михаил", "Ирина", "Никита", "Светлана", "Павел", "Юлия",
	"Артём", "Екатерина", "Роман", "Анна", "Максим", "Ксения", "Егор", "Вера",
}

var lastNames = []string{
	"Иванов", "Смирнова", "Кузнецов", "Попова", "Соколов", "Лебедева", "Новиков", "Козлова",
	"Морозов", "Петрова", "Волков", "Соловьёва", "Зайцев", "Павлова", "Семёнов", "Голубева",
	"Виноградов", "Богданова", "Воробьёв", "Фёдорова", "Михайлов", "Белова", "Тарасов", "Комарова",
}

func demoPersonas() []guestSpec {
	return []guestSpec{
		{
			Key: "g-anna", Name: "Анна Соколова", Status: "active", CreatedDaysAgo: 150,
			Identifiers: []identifierSpec{
				{"phone", demoPhone(1), "pos-bean"},
				{"email", demoEmail(1), "crm-cinema"},
			},
			Events: []eventSpec{
				{GuestKey: "g-anna", Tenant: "cafe-bean", Type: "purchase", Source: "pos-bean", DaysAgo: 40, Hour: 9, Minute: 15, Amount: 42000, HasAmount: true, ExternalID: "CHK-100001", Keys: []string{"phone"}},
				{GuestKey: "g-anna", Tenant: "cinema-sinema", Project: "imax", Type: "purchase", Source: "pos-cinema", DaysAgo: 25, Hour: 20, Minute: 5, Amount: 180000, HasAmount: true, ExternalID: "CHK-100002", Keys: []string{"phone", "email"}},
				{GuestKey: "g-anna", Tenant: "cafe-bean", Type: "purchase", Source: "pos-bean", DaysAgo: 8, Hour: 8, Minute: 50, Amount: 36000, HasAmount: true, ExternalID: "CHK-100003", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-olga", Name: "Ольга Петрова", Status: "active", CreatedDaysAgo: 120,
			Identifiers: []identifierSpec{{"phone", demoPhone(2), "pos-bean"}},
			Events: []eventSpec{
				{GuestKey: "g-olga", Tenant: "cafe-bean", Type: "purchase", Source: "pos-bean", DaysAgo: 33, Hour: 12, Minute: 30, Amount: 28000, HasAmount: true, ExternalID: "CHK-100010", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-igor", Name: "Игорь Сидоров", Status: "active", CreatedDaysAgo: 110,
			Identifiers: []identifierSpec{{"phone", demoPhone(2), "crm-cinema"}},
			Events: []eventSpec{
				{GuestKey: "g-igor", Tenant: "cinema-sinema", Project: "vip", Type: "purchase", Source: "pos-cinema", DaysAgo: 31, Hour: 19, Minute: 40, Amount: 240000, HasAmount: true, ExternalID: "CHK-100011", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-sleep-1", Name: "Виктор Ершов", Status: "active", CreatedDaysAgo: 170,
			Identifiers: []identifierSpec{{"phone", demoPhone(3), "pos-cinema"}},
			Events: []eventSpec{
				{GuestKey: "g-sleep-1", Tenant: "cinema-sinema", Project: "imax", Type: "purchase", Source: "pos-cinema", DaysAgo: 20, Hour: 21, Minute: 0, Amount: 200000, HasAmount: true, ExternalID: "CHK-100020", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-sleep-2", Name: "Дарья Крылова", Status: "active", CreatedDaysAgo: 165,
			Identifiers: []identifierSpec{{"phone", demoPhone(4), "pos-cinema"}},
			Events: []eventSpec{
				{GuestKey: "g-sleep-2", Tenant: "cinema-sinema", Project: "kids-hall", Type: "purchase", Source: "pos-cinema", DaysAgo: 15, Hour: 13, Minute: 20, Amount: 120000, HasAmount: true, ExternalID: "CHK-100021", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-sleep-3", Name: "Пётр Лавров", Status: "active", CreatedDaysAgo: 160,
			Identifiers: []identifierSpec{{"phone", demoPhone(5), "pos-cinema"}},
			Events: []eventSpec{
				{GuestKey: "g-sleep-3", Tenant: "cinema-sinema", Project: "vip", Type: "purchase", Source: "pos-cinema", DaysAgo: 12, Hour: 20, Minute: 45, Amount: 260000, HasAmount: true, ExternalID: "CHK-100022", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-sleep-4", Name: "Марина Гусева", Status: "active", CreatedDaysAgo: 155,
			Identifiers: []identifierSpec{{"email", demoEmail(4), "crm-cinema"}},
			Events: []eventSpec{
				{GuestKey: "g-sleep-4", Tenant: "cinema-sinema", Project: "imax", Type: "booking", Source: "crm-cinema", DaysAgo: 9, Hour: 18, Minute: 10, Amount: 150000, HasAmount: true, ExternalID: "BK-100023", Keys: []string{"email"}},
			},
		},
		{
			Key: "g-sleep-5", Name: "Артём Белов", Status: "active", CreatedDaysAgo: 150,
			Identifiers: []identifierSpec{{"phone", demoPhone(6), "pos-cinema"}},
			Events: []eventSpec{
				{GuestKey: "g-sleep-5", Tenant: "cinema-sinema", Project: "imax", Type: "purchase", Source: "pos-cinema", DaysAgo: 6, Hour: 19, Minute: 30, Amount: 190000, HasAmount: true, ExternalID: "CHK-100024", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-sms", Name: "Светлана Орлова", Status: "active", CreatedDaysAgo: 140,
			Identifiers: []identifierSpec{{"phone", demoPhone(7), "pos-cinema"}},
			Events: []eventSpec{
				{GuestKey: "g-sms", Tenant: "cinema-sinema", Project: "imax", Type: "purchase", Source: "pos-cinema", DaysAgo: 5, Hour: 20, Minute: 0, Amount: 170000, HasAmount: true, ExternalID: "CHK-100030", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-new", Name: "Никита Захаров", Status: "active", CreatedDaysAgo: 3,
			Identifiers: []identifierSpec{{"phone", demoPhone(8), "pos-bean"}},
		},
		{
			Key: "g-vip", Name: "Екатерина Волкова", Status: "active", CreatedDaysAgo: 175,
			Identifiers: []identifierSpec{
				{"phone", demoPhone(9), "pos-cinema"},
				{"email", demoEmail(9), "crm-cinema"},
				{"card", demoCard(9), "pos-trend"},
			},
			Events: []eventSpec{
				{GuestKey: "g-vip", Tenant: "cinema-sinema", Project: "vip", Type: "purchase", Source: "pos-cinema", DaysAgo: 50, Hour: 20, Minute: 15, Amount: 480000, HasAmount: true, ExternalID: "CHK-100040", Keys: []string{"phone", "email"}},
				{GuestKey: "g-vip", Tenant: "fashion-trend", Type: "purchase", Source: "pos-trend", DaysAgo: 44, Hour: 15, Minute: 5, Amount: 460000, HasAmount: true, ExternalID: "CHK-100041", Keys: []string{"card"}},
				{GuestKey: "g-vip", Tenant: "food-plate", Project: "banquet", Type: "purchase", Source: "csv-food", DaysAgo: 30, Hour: 19, Minute: 0, Amount: 420000, HasAmount: true, ExternalID: "CHK-100042", Keys: []string{"phone"}},
				{GuestKey: "g-vip", Tenant: "beauty-lokon", Type: "purchase", Source: "pos-trend", DaysAgo: 14, Hour: 12, Minute: 40, Amount: 380000, HasAmount: true, ExternalID: "CHK-100043", Keys: []string{"card"}},
			},
		},
		{
			Key: "g-churn", Name: "Роман Тихонов", Status: "active", CreatedDaysAgo: 180,
			Identifiers: []identifierSpec{{"phone", demoPhone(10), "crm-fit"}},
			Events: []eventSpec{
				{GuestKey: "g-churn", Tenant: "fit-zone", Project: "gym", Type: "purchase", Source: "crm-fit", DaysAgo: 120, Hour: 7, Minute: 30, Amount: 350000, HasAmount: true, ExternalID: "CHK-100050", Keys: []string{"phone"}},
				{GuestKey: "g-churn", Tenant: "fit-zone", Project: "pool", Type: "visit", Source: "crm-fit", DaysAgo: 60, Hour: 8, Minute: 0, ExternalID: "VIS-100051", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-email", Name: "Ирина Мельникова", Status: "active", CreatedDaysAgo: 130,
			Identifiers: []identifierSpec{{"email", demoEmail(11), "crm-cinema"}},
			Events: []eventSpec{
				{GuestKey: "g-email", Tenant: "cinema-sinema", Project: "imax", Type: "booking", Source: "crm-cinema", DaysAgo: 18, Hour: 17, Minute: 25, Amount: 130000, HasAmount: true, ExternalID: "BK-100060", Keys: []string{"email"}},
			},
		},
		{
			Key: "g-card", Name: "Максим Дроздов", Status: "active", CreatedDaysAgo: 125,
			Identifiers: []identifierSpec{{"card", demoCard(12), "pos-trend"}},
			Events: []eventSpec{
				{GuestKey: "g-card", Tenant: "fashion-trend", Type: "purchase", Source: "pos-trend", DaysAgo: 22, Hour: 16, Minute: 10, Amount: 210000, HasAmount: true, ExternalID: "CHK-100070", Keys: []string{"card"}},
			},
		},
		{
			Key: "g-reversal", Name: "Юлия Барсукова", Status: "active", CreatedDaysAgo: 115,
			Identifiers: []identifierSpec{{"phone", demoPhone(13), "pos-bean"}},
			Events: []eventSpec{
				{GuestKey: "g-reversal", Tenant: "cafe-bean", Type: "purchase", Source: "pos-bean", DaysAgo: 27, Hour: 11, Minute: 5, Amount: 95000, HasAmount: true, ExternalID: "CHK-100080", Keys: []string{"phone"}},
				{GuestKey: "g-reversal", Tenant: "cafe-bean", Type: "reversal", Source: "pos-bean", DaysAgo: 26, Hour: 10, Minute: 20, Amount: 95000, HasAmount: true, ExternalID: "REV-100081", Reverses: "CHK-100080", Keys: []string{"phone"}},
			},
		},
		{
			Key: "g-topup", Name: "Егор Панкратов", Status: "active", CreatedDaysAgo: 105,
			Identifiers: []identifierSpec{{"phone", demoPhone(14), "crm-fit"}},
			Events: []eventSpec{
				{GuestKey: "g-topup", Tenant: "fit-zone", Project: "gym", Type: "top_up", Source: "crm-fit", DaysAgo: 35, Hour: 9, Minute: 0, Amount: 500000, HasAmount: true, ExternalID: "TOP-100090", Keys: []string{"phone"}},
				{GuestKey: "g-topup", Tenant: "fit-zone", Project: "gym", Type: "spend", Source: "crm-fit", DaysAgo: 21, Hour: 18, Minute: 30, Amount: 120000, HasAmount: true, ExternalID: "USE-100091", Keys: []string{"phone"}},
			},
		},
	}
}

func demoMergedGuests() []guestSpec {
	out := make([]guestSpec, 0, 8)
	for i := 1; i <= 8; i++ {
		out = append(out, guestSpec{
			Key:            fmt.Sprintf("g-merged-%d", i),
			Name:           fmt.Sprintf("Дубликат %d", i),
			Status:         "merged",
			MergedInto:     backgroundGuestKey(i * 3),
			CreatedDaysAgo: 90 + i,
		})
	}
	return out
}

func demoDeletedGuests() []guestSpec {
	out := make([]guestSpec, 0, 4)
	for i := 1; i <= 4; i++ {
		out = append(out, guestSpec{
			Key:            fmt.Sprintf("g-deleted-%d", i),
			Name:           fmt.Sprintf("Удалённый профиль %d", i),
			Status:         "deleted",
			CreatedDaysAgo: 100 + i,
		})
	}
	return out
}

func backgroundGuest(i int) guestSpec {
	r := newRNG(uint64(i)*guestSeedSalt + 7919)
	g := guestSpec{
		Key:            backgroundGuestKey(i),
		Status:         "active",
		CreatedDaysAgo: 30 + r.intn(150),
	}
	if r.chance(85) {
		g.Name = firstNames[r.intn(len(firstNames))] + " " + lastNames[r.intn(len(lastNames))]
	}
	if r.chance(90) {
		g.Identifiers = append(g.Identifiers, identifierSpec{"phone", demoPhone(100 + i), "pos-bean"})
	}
	if r.chance(40) {
		g.Identifiers = append(g.Identifiers, identifierSpec{"email", demoEmail(100 + i), "crm-cinema"})
	}
	if r.chance(15) {
		g.Identifiers = append(g.Identifiers, identifierSpec{"card", demoCard(100 + i), "pos-trend"})
	}
	if len(g.Identifiers) == 0 {
		g.Identifiers = append(g.Identifiers, identifierSpec{"phone", demoPhone(100 + i), "pos-bean"})
	}

	tenants := backgroundTenants(i, r)
	eventCount := 8 + r.intn(14)
	for e := 0; e < eventCount; e++ {
		tenant := tenants[r.intn(len(tenants))]
		daysAgo := r.intn(180)
		if r.chance(60) {
			daysAgo = r.intn(60)
		}
		ev := eventSpec{
			GuestKey:   g.Key,
			Tenant:     tenant,
			DaysAgo:    daysAgo,
			Hour:       8 + r.intn(14),
			Minute:     r.intn(60),
			ExternalID: fmt.Sprintf("BG-%05d-%02d", i, e),
			Keys:       identifierKeys(g.Identifiers),
		}
		ev.Type, ev.Source, ev.Project = backgroundEventShape(tenant, r)
		if ev.Type == "purchase" || ev.Type == "booking" {
			ev.HasAmount = true
			ev.Amount = int64(15000 + r.intn(485000))
		}
		g.Events = append(g.Events, ev)
	}
	return g
}

func identifierKeys(ids []identifierSpec) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.Type)
	}
	return out
}

func backgroundTenants(i int, r *rng) []string {
	switch i % 10 {
	case 0, 1:
		return []string{"cinema-sinema"}
	case 2, 3:
		return []string{"cafe-bean"}
	case 4:
		return []string{"cafe-bean", "cinema-sinema"}
	case 5:
		return []string{"fit-zone", "food-plate"}
	case 6:
		return []string{"cinema-sinema", "food-plate"}
	case 7:
		return []string{"beauty-lokon", "fashion-trend"}
	case 8:
		return []string{"kids-cube", "books-line"}
	default:
		pool := []string{"cafe-bean", "cinema-sinema", "fit-zone", "food-plate", "kids-cube", "beauty-lokon", "fashion-trend", "tech-volt", "books-line", "pets-tail"}
		first := pool[r.intn(len(pool))]
		second := pool[r.intn(len(pool))]
		if second == first {
			return []string{first}
		}
		return []string{first, second}
	}
}

func backgroundEventShape(tenant string, r *rng) (eventType, source, project string) {
	switch tenant {
	case "cafe-bean":
		source = "pos-bean"
	case "cinema-sinema":
		source = "pos-cinema"
		projects := []string{"imax", "vip", "kids-hall"}
		project = projects[r.intn(len(projects))]
	case "fit-zone":
		source = "crm-fit"
		if r.chance(50) {
			project = "gym"
		} else {
			project = "pool"
		}
	case "food-plate":
		source = "csv-food"
		if r.chance(30) {
			project = "banquet"
		} else {
			project = "main"
		}
	case "kids-cube":
		source = "api-kids"
		project = "main"
	case "fashion-trend", "beauty-lokon":
		source = "pos-trend"
	default:
		source = "wifi-mall"
	}
	types := []string{"purchase", "visit", "purchase", "visit", "purchase", "booking"}
	return types[r.intn(len(types))], source, project
}

func (s *Seeder) buildGuests() []guestSpec {
	out := make([]guestSpec, 0, s.opts.Guests+len(demoPersonas())+12)
	for i := 0; i < s.opts.Guests; i++ {
		out = append(out, backgroundGuest(i))
	}
	out = append(out, demoPersonas()...)
	out = append(out, demoMergedGuests()...)
	out = append(out, demoDeletedGuests()...)
	return out
}

func (s *Seeder) buildEvents() []eventSpec {
	out := make([]eventSpec, 0, 1600)
	for _, g := range s.buildGuests() {
		out = append(out, g.Events...)
	}
	return append(out, unmatchedEvents()...)
}
