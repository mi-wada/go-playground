package main

import (
	"fmt"
	"net/http"
	"sync"
)

type Todo struct {
	ID   int
	Text string
}

var (
	todos   = []Todo{}
	todosMu sync.Mutex
	nextID  = 0
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ToDo App</title>
    <script src="https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js" integrity="sha384-/TgkGk7p307TH7EXJDuUlgG3Ce1UVolAOFopFekQkkXihi5u/6OCvVKyz1W+idaz" crossorigin="anonymous"></script>
</head>
<body>
    <h1>ToDo</h1>

    <form hx-post="/todos" hx-target="#todo-list" hx-swap="beforeend" hx-on::after-request="this.reset()">
        <input type="text" name="todo" placeholder="新しいタスク" required>
        <button type="submit">追加</button>
    </form>

    <ul id="todo-list" hx-get="/todos" hx-trigger="load"></ul>
</body>
</html>
		`))
	})

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.Header().Set("Content-Type", "text/html")
			text := r.FormValue("todo")

			todosMu.Lock()
			todo := Todo{ID: nextID, Text: text}
			nextID++
			todos = append(todos, todo)
			todosMu.Unlock()

			fmt.Fprintf(w, `
<li id="todo-%d">
    <span>%s</span>
    <button hx-get="/todos/%d/edit" hx-target="#todo-%d" hx-swap="outerHTML">編集</button>
    <button hx-delete="/todos/%d" hx-target="#todo-%d" hx-swap="outerHTML">削除</button>
</li>
`, todo.ID, todo.Text, todo.ID, todo.ID, todo.ID, todo.ID)

		case http.MethodGet:
			w.Header().Set("Content-Type", "text/html")
			todosMu.Lock()
			defer todosMu.Unlock()

			for _, todo := range todos {
				fmt.Fprintf(w, `
<li id="todo-%d">
    <span>%s</span>
    <button hx-get="/todos/%d/edit" hx-target="#todo-%d" hx-swap="outerHTML">編集</button>
    <button hx-delete="/todos/%d" hx-target="#todo-%d" hx-swap="outerHTML">削除</button>
</li>
`, todo.ID, todo.Text, todo.ID, todo.ID, todo.ID, todo.ID)
			}
		}
	})

	http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		var id int

		// 編集フォーム表示
		if n, _ := fmt.Sscanf(r.URL.Path, "/todos/%d/edit", &id); n == 1 && r.Method == "GET" {
			w.Header().Set("Content-Type", "text/html")
			todosMu.Lock()
			var todo *Todo
			for i := range todos {
				if todos[i].ID == id {
					todo = &todos[i]
					break
				}
			}
			todosMu.Unlock()

			if todo == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			fmt.Fprintf(w, `
<li id="todo-%d">
    <form hx-put="/todos/%d" hx-target="#todo-%d" hx-swap="outerHTML">
        <input type="text" name="todo" value="%s" required>
        <button type="submit">保存</button>
        <button type="button" hx-get="/todos" hx-target="#todo-list" hx-swap="innerHTML">キャンセル</button>
    </form>
</li>
`, todo.ID, todo.ID, todo.ID, todo.Text)
			return
		}

		// Todo更新
		if n, _ := fmt.Sscanf(r.URL.Path, "/todos/%d", &id); n == 1 && r.Method == "PUT" {
			w.Header().Set("Content-Type", "text/html")
			r.ParseForm()
			text := r.FormValue("todo")

			todosMu.Lock()
			var todo *Todo
			for i := range todos {
				if todos[i].ID == id {
					todos[i].Text = text
					todo = &todos[i]
					break
				}
			}
			todosMu.Unlock()

			if todo == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			fmt.Fprintf(w, `
<li id="todo-%d">
    <span>%s</span>
    <button hx-get="/todos/%d/edit" hx-target="#todo-%d" hx-swap="outerHTML">編集</button>
    <button hx-delete="/todos/%d" hx-target="#todo-%d" hx-swap="outerHTML">削除</button>
</li>
`, todo.ID, todo.Text, todo.ID, todo.ID, todo.ID, todo.ID)
			return
		}

		// Todo削除
		if n, _ := fmt.Sscanf(r.URL.Path, "/todos/%d", &id); n == 1 && r.Method == "DELETE" {
			todosMu.Lock()
			for i := range todos {
				if todos[i].ID == id {
					todos = append(todos[:i], todos[i+1:]...)
					break
				}
			}
			todosMu.Unlock()

			w.WriteHeader(http.StatusOK)
		}
	})

	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		panic(err)
	}
}
