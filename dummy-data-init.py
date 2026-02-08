import datetime
import random
import sqlite3

# Align with Go's InitDB path
DB_PATH = "src/data/state.db"

# Define dummy users
users = [
    {"email": "alice@example.com", "username": "alice"},
    {"email": "bob@example.com", "username": "bob"},
    {"email": "carol@example.com", "username": "carol"},
    {"email": "dave@example.com", "username": "dave"},
    {"email": "eve@example.com", "username": "eve"},
]

# Tags that can be used for events
tags = ["weight", "pushups", "exercise"]


# Helper to generate a random datetime between two dates
def random_datetime(start, end):
    delta = end - start
    random_seconds = random.randint(0, int(delta.total_seconds()))
    return start + datetime.timedelta(seconds=random_seconds)


# Date range: Jan 1 2026 to July 31 2026
start_date = datetime.datetime(2026, 1, 1)
end_date = datetime.datetime(2026, 7, 31, 23, 59, 59)


def main():
    conn = sqlite3.connect(DB_PATH)
    cur = conn.cursor()

    # Ensure tables exist (schema mirrors database.go)
    cur.execute("""
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT,
            email TEXT
        );
    """)
    cur.execute("""
        CREATE TABLE IF NOT EXISTS events (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            tag TEXT,
            comment TEXT,
            value TEXT,
            recordedAt TEXT,
            recordedBy TEXT
        );
    """)

    # Insert users (id is auto‑generated)
    cur.executemany(
        "INSERT INTO users (username, email) VALUES (?, ?);",
        [(u["username"], u["email"]) for u in users],
    )

    # Generate and insert events
    event_rows = []
    for user in users:
        for _ in range(100):
            tag = random.choice(tags)

            if tag == "exercise":
                value = ""
                comment = random.choice(
                    [
                        "Morning yoga session",
                        "Evening run",
                        "Cycling in the park",
                        "Swimming laps",
                        "Group fitness class",
                    ]
                )
            else:
                # weight in kg or pushups count, stored as string
                if tag == "weight":
                    value = str(random.randint(50, 120))  # kg
                else:  # pushups
                    value = str(random.randint(10, 100))

                comment = random.choice(
                    [
                        f"{tag.capitalize()} recorded",
                        "",
                        f"Felt good after {tag}",
                        f"Today's {tag} progress",
                    ]
                )

            recorded_at = random_datetime(start_date, end_date).isoformat(
                sep=" ", timespec="seconds"
            )

            event_rows.append(
                (
                    tag,
                    comment if comment else "",
                    value,
                    recorded_at,
                    user["email"],
                )
            )

    cur.executemany(
        """
        INSERT INTO events (tag, comment, value, recordedAt, recordedBy)
        VALUES (?, ?, ?, ?, ?);
        """,
        event_rows,
    )

    conn.commit()
    conn.close()
    print(f"Inserted {len(users)} users and {len(event_rows)} events into {DB_PATH}")


if __name__ == "__main__":
    main()
