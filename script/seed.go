package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lzimin05/course-todo/config"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к базе данных
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.DB,
	)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Запускаем заполнение тестовыми данными
	if err := seedDatabase(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("Database seeded successfully!")
}

func seedDatabase(db *sql.DB) error {
	log.Println("Starting to seed database...")

	// Создаем тестовых пользователей
	users, err := createTestUsers(db)
	if err != nil {
		return fmt.Errorf("failed to create test users: %w", err)
	}
	log.Printf("Created %d test users", len(users))

	// Создаем тестовые проекты
	projects, err := createTestProjects(db, users)
	if err != nil {
		return fmt.Errorf("failed to create test projects: %w", err)
	}
	log.Printf("Created %d test projects", len(projects))

	// Добавляем участников в проекты
	if err := createProjectMembers(db, projects, users); err != nil {
		return fmt.Errorf("failed to create project members: %w", err)
	}
	log.Println("Added project members")

	// Создаем тестовые задачи
	tasks, err := createTestTasks(db, projects, users)
	if err != nil {
		return fmt.Errorf("failed to create test tasks: %w", err)
	}
	log.Printf("Created %d test tasks", len(tasks))

	// Создаем тестовые заметки
	notes, err := createTestNotes(db, projects, users)
	if err != nil {
		return fmt.Errorf("failed to create test notes: %w", err)
	}
	log.Printf("Created %d test notes", len(notes))

	return nil
}

type TestUser struct {
	ID       uuid.UUID
	Login    string
	Username string
	Email    string
}

func createTestUsers(db *sql.DB) ([]TestUser, error) {
	users := []TestUser{
		{Login: "test_admin", Username: "Администратор", Email: "admin@test.com"},
		{Login: "test_john", Username: "Джон Доу", Email: "john@test.com"},
		{Login: "test_jane", Username: "Джейн Смит", Email: "jane@test.com"},
		{Login: "test_alice", Username: "Алиса Иванова", Email: "alice@test.com"},
		{Login: "test_bob", Username: "Боб Петров", Email: "bob@test.com"},
		{Login: "test_charlie", Username: "Чарли Сидоров", Email: "charlie@test.com"},
	}

	// Хешируем пароль "password123" для всех тестовых пользователей
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO todo."user" (id, login, username, email, password_hash)
		VALUES ($1, $2, $3, $4, $5)
	`

	for i := range users {
		users[i].ID = uuid.New()
		_, err := db.Exec(query, users[i].ID, users[i].Login, users[i].Username, users[i].Email, passwordHash)
		if err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", users[i].Login, err)
		}
	}

	return users, nil
}

type TestProject struct {
	ID          uuid.UUID
	Name        string
	Description string
	OwnerID     uuid.UUID
}

func createTestProjects(db *sql.DB, users []TestUser) ([]TestProject, error) {
	projects := []TestProject{
		{
			Name:        "Разработка мобильного приложения",
			Description: "Создание мобильного приложения для управления задачами",
			OwnerID:     users[0].ID, // admin
		},
		{
			Name:        "Веб-сайт компании",
			Description: "Разработка корпоративного веб-сайта",
			OwnerID:     users[1].ID, // john
		},
		{
			Name:        "Система аналитики",
			Description: "Внедрение системы аналитики данных",
			OwnerID:     users[2].ID, // jane
		},
		{
			Name:        "Личный проект",
			Description: "Разработка личного портфолио",
			OwnerID:     users[3].ID, // alice
		},
		{
			Name:        "Open Source библиотека",
			Description: "Создание библиотеки для работы с API",
			OwnerID:     users[4].ID, // bob
		},
	}

	query := `
		INSERT INTO todo.project (id, name, description, owner_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	for i := range projects {
		projects[i].ID = uuid.New()
		_, err := db.Exec(query, 
			projects[i].ID, 
			projects[i].Name, 
			projects[i].Description, 
			projects[i].OwnerID, 
			time.Now().Add(-time.Duration(i*24)*time.Hour), // разные даты создания
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create project %s: %w", projects[i].Name, err)
		}
	}

	return projects, nil
}

func createProjectMembers(db *sql.DB, projects []TestProject, users []TestUser) error {
	query := `
		INSERT INTO todo.project_member (id, project_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	// Добавляем владельцев проектов как участников с ролью owner
	for _, project := range projects {
		memberID := uuid.New()
		_, err := db.Exec(query, memberID, project.ID, project.OwnerID, "owner", time.Now())
		if err != nil {
			return fmt.Errorf("failed to add owner as member: %w", err)
		}
	}

	// Добавляем других участников в проекты
	members := []struct {
		ProjectIndex int
		UserIndex    int
	}{
		{0, 1}, {0, 2}, {0, 3}, // В первый проект добавляем john, jane, alice
		{1, 0}, {1, 4}, {1, 5}, // Во второй проект добавляем admin, bob, charlie
		{2, 0}, {2, 1}, {2, 4}, // В третий проект добавляем admin, john, bob
		{3, 1}, {3, 2},         // В четвертый проект добавляем john, jane
		{4, 0}, {4, 2}, {4, 3}, // В пятый проект добавляем admin, jane, alice
	}

	for _, member := range members {
		if member.UserIndex < len(users) && member.ProjectIndex < len(projects) {
			memberID := uuid.New()
			_, err := db.Exec(query, 
				memberID, 
				projects[member.ProjectIndex].ID, 
				users[member.UserIndex].ID, 
				"member", 
				time.Now().Add(-time.Duration(member.ProjectIndex*12)*time.Hour),
			)
			if err != nil {
				return fmt.Errorf("failed to add project member: %w", err)
			}
		}
	}

	return nil
}

type TestTask struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Importance  int
	Status      string
	Deadline    time.Time
}

func createTestTasks(db *sql.DB, projects []TestProject, users []TestUser) ([]TestTask, error) {
	
	tasks := []TestTask{
		// Задачи для первого проекта
		{
			ProjectID:   projects[0].ID,
			UserID:      users[1].ID,
			Title:       "Дизайн пользовательского интерфейса",
			Description: "Создать макеты основных экранов приложения",
			Importance:  5,
			Status:      "in_progress",
			Deadline:    time.Now().Add(7 * 24 * time.Hour),
		},
		{
			ProjectID:   projects[0].ID,
			UserID:      users[2].ID,
			Title:       "Настройка базы данных",
			Description: "Создать схему БД и настроить подключение",
			Importance:  4,
			Status:      "completed",
			Deadline:    time.Now().Add(-2 * 24 * time.Hour),
		},
		{
			ProjectID:   projects[0].ID,
			UserID:      users[3].ID,
			Title:       "Разработка API",
			Description: "Создать REST API для мобильного приложения",
			Importance:  5,
			Status:      "waiting",
			Deadline:    time.Now().Add(14 * 24 * time.Hour),
		},
		
		// Задачи для второго проекта
		{
			ProjectID:   projects[1].ID,
			UserID:      users[0].ID,
			Title:       "Создание главной страницы",
			Description: "Разработать главную страницу сайта",
			Importance:  3,
			Status:      "in_progress",
			Deadline:    time.Now().Add(5 * 24 * time.Hour),
		},
		{
			ProjectID:   projects[1].ID,
			UserID:      users[4].ID,
			Title:       "Интеграция с CMS",
			Description: "Настроить систему управления контентом",
			Importance:  2,
			Status:      "waiting",
			Deadline:    time.Now().Add(21 * 24 * time.Hour),
		},
		
		// Задачи для третьего проекта
		{
			ProjectID:   projects[2].ID,
			UserID:      users[0].ID,
			Title:       "Анализ требований",
			Description: "Проанализировать бизнес-требования для системы аналитики",
			Importance:  4,
			Status:      "completed",
			Deadline:    time.Now().Add(-5 * 24 * time.Hour),
		},
		{
			ProjectID:   projects[2].ID,
			UserID:      users[1].ID,
			Title:       "Выбор технологий",
			Description: "Выбрать подходящие технологии для реализации",
			Importance:  3,
			Status:      "in_progress",
			Deadline:    time.Now().Add(3 * 24 * time.Hour),
		},
		
		// Задачи для четвертого проекта
		{
			ProjectID:   projects[3].ID,
			UserID:      users[1].ID,
			Title:       "Создание портфолио",
			Description: "Разработать дизайн персонального сайта",
			Importance:  2,
			Status:      "waiting",
			Deadline:    time.Now().Add(30 * 24 * time.Hour),
		},
		
		// Задачи для пятого проекта
		{
			ProjectID:   projects[4].ID,
			UserID:      users[0].ID,
			Title:       "Документация API",
			Description: "Написать документацию для библиотеки",
			Importance:  3,
			Status:      "in_progress",
			Deadline:    time.Now().Add(10 * 24 * time.Hour),
		},
		{
			ProjectID:   projects[4].ID,
			UserID:      users[2].ID,
			Title:       "Написание тестов",
			Description: "Создать unit тесты для основного функционала",
			Importance:  4,
			Status:      "waiting",
			Deadline:    time.Now().Add(15 * 24 * time.Hour),
		},
	}

	query := `
		INSERT INTO todo."task" (id, project_id, user_id, title, description, importance, status, created_at, deadline)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	for i := range tasks {
		tasks[i].ID = uuid.New()
		createdAt := time.Now().Add(-time.Duration(i*6) * time.Hour)
		
		_, err := db.Exec(query,
			tasks[i].ID,
			tasks[i].ProjectID,
			tasks[i].UserID,
			tasks[i].Title,
			tasks[i].Description,
			tasks[i].Importance,
			tasks[i].Status,
			createdAt,
			tasks[i].Deadline,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create task %s: %w", tasks[i].Title, err)
		}
	}

	return tasks, nil
}

type TestNote struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
}

func createTestNotes(db *sql.DB, projects []TestProject, users []TestUser) ([]TestNote, error) {
	notes := []TestNote{
		{
			ProjectID:   projects[0].ID,
			UserID:      users[0].ID,
			Name:        "Идеи для UI",
			Description: "Коллекция идей для улучшения пользовательского интерфейса:\n- Использовать темную тему\n- Добавить анимации переходов\n- Реализовать drag&drop",
		},
		{
			ProjectID:   projects[0].ID,
			UserID:      users[1].ID,
			Name:        "Технические требования",
			Description: "Основные технические требования к проекту:\n- Поддержка iOS и Android\n- Оффлайн режим\n- Синхронизация данных",
		},
		{
			ProjectID:   projects[1].ID,
			UserID:      users[1].ID,
			Name:        "Структура сайта",
			Description: "Планируемая структура веб-сайта:\n1. Главная страница\n2. О компании\n3. Услуги\n4. Контакты\n5. Блог",
		},
		{
			ProjectID:   projects[1].ID,
			UserID:      users[4].ID,
			Name:        "SEO оптимизация",
			Description: "Чек-лист для SEO:\n- Мета-теги\n- Структурированные данные\n- Оптимизация изображений\n- Карта сайта",
		},
		{
			ProjectID:   projects[2].ID,
			UserID:      users[2].ID,
			Name:        "Метрики для отслеживания",
			Description: "Ключевые метрики системы аналитики:\n- Конверсия\n- Время на сайте\n- Количество страниц за сессию\n- Источники трафика",
		},
		{
			ProjectID:   projects[3].ID,
			UserID:      users[3].ID,
			Name:        "Проекты для портфолио",
			Description: "Список проектов для включения в портфолио:\n- Todo приложение\n- E-commerce сайт\n- Блог на React\n- Мобильное приложение",
		},
		{
			ProjectID:   projects[4].ID,
			UserID:      users[4].ID,
			Name:        "Roadmap библиотеки",
			Description: "Планы развития библиотеки:\nv1.0 - Базовый функционал\nv1.1 - Кэширование запросов\nv1.2 - Поддержка GraphQL\nv2.0 - Typescript поддержка",
		},
		{
			ProjectID:   projects[4].ID,
			UserID:      users[0].ID,
			Name:        "Примеры использования",
			Description: "Примеры кода для документации:\n```javascript\nconst api = new APILib();\napi.get('/users').then(data => console.log(data));\n```",
		},
	}

	query := `
		INSERT INTO todo.note (id, project_id, user_id, name, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	for i := range notes {
		notes[i].ID = uuid.New()
		createdAt := time.Now().Add(-time.Duration(i*4) * time.Hour)
		
		_, err := db.Exec(query,
			notes[i].ID,
			notes[i].ProjectID,
			notes[i].UserID,
			notes[i].Name,
			notes[i].Description,
			createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create note %s: %w", notes[i].Name, err)
		}
	}

	return notes, nil
}
