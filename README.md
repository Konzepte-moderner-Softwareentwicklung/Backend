[![SonarQube](https://scm.thm.de/sonar/api/project_badges/measure?project=kmsbackend&metric=alert_status)](https://scm.thm.de/sonar/dashboard?id=kmsbackend)

# ![Logo](https://github.com/Konzepte-moderner-Softwareentwicklung/Backend/blob/f27c180dcd0b50ee5533e5bdb1ae97030adead33/readme-content/Logo-smaller.png?raw=true)
# Microservices Backend (Entwicklungsumgebung)

## Übersicht

Dieses Backend stellt eine Microservices-Architektur bereit, die mittels Docker Compose orchestriert wird.
Die Architektur umfasst folgende Services:

- **NATS** (Message Broker mit JetStream)
- **MongoDB** (Datenbank)
- **MinIO** (Objektspeicher)
- **Elasticsearch, Logstash & Kibana** (Logging & Monitoring)
- **Verschiedene Anwendungsservices** (User, Angebot, Rating, Tracking, Media, Chat, Gateway)
- **NATS-UI** (Monitoring-Interface für NATS)
- **nginx** als Reverse Proxy für das Frontend und TLS-Termination

**Wichtig:** Diese Docker Compose Konfiguration ist ausschließlich für Entwicklungszwecke gedacht!

---

## Docker Compose Services

| Service           | Beschreibung                               | Port(s)           |
|------------------|---------------------------------------------|-------------------|
| nats             | Message Broker mit JetStream                | 4222, 8222        |
| nats-ui          | NATS Monitoring UI                          | 31311             |
| mongo            | MongoDB Datenbank                           | 27017             |
| minio            | Objektspeicher (S3-kompatibel)              | 9000, 9001        |
| gateway          | API Gateway                                 | 8081              |
| user-service     | Benutzerverwaltung                          | 8082              |
| angebot-service  | Verwaltung von Angeboten und Gesuchen       | 8084              |
| tracking-service | Tracking der Fahrten                        | 8085              |
| media-service    | Verwaltung von Medieninhalten               | 8083              |
| rating-service   | Bewertungssystem                            | -                 |
| chat-service     | Chat-Funktion zwischen Nutzern              | -                 |
| frontend         | Frontend Webanwendung                       | 8080              |
| nginx            | Reverse Proxy & TLS                         | 80, 443           |
| elasticsearch    | Speichert strukturierte Log-Daten           | 9200              |
| kibana           | Analyse-UI für Logs                         | 5601              |
| logstash         | Verarbeitung & Weiterleitung von Logs       | -                 |

---

## Features

### Registrierung
- Pflichtfelder: Vorname, Nachname, E-Mail (zweimal), Passwort, Geburtstag (ab 18 Jahren)
- Zusätzliche Felder bei Angebotserstellung: Handynummer (privat), Profilbild

### Login
- E-Mail und Passwort oder mit E-Mail und Webauthn (Passkey, apple FaceID, Fingerabdruck)

### Suche
- Suche nach Angeboten oder Gesuchen
- Filtermöglichkeiten: Zeitraum (Von/Bis), Fracht (Gewicht/Maße), Bewertung, verfügbare Plätze

### Profilansicht (registrierte Benutzer)
- Öffentlich: Vorname, Nachname (nur erster Buchstabe), Profilbild, Alter, Notizen
- Bewertungen, Anzahl der Fahrten (Angeboten/Gesucht)
- Erfahrung (Mitfahrer, Frachtgewicht, Strecke, Sprachen, Raucherstatus)

### Benutzer- & Fahrzeugverwaltung
- Profil und Fahrzeugdaten editierbar
- Fahrzeugattribute: Gewicht, Maße, Sonderfunktionen (z.B. Kühlung)

### Bewertungen
- 5-Sterne-Skala nach erfolgter Fahrt
- Gegenseitige Bewertung Fahrer <-> Mitfahrer
- Bewertung nur bei nicht stornierten Fahrten möglich
- Fragen z.B. Pünktlichkeit, Einhaltung Abmachungen, Wohlfühlen, Frachtzustand

### Angebot/Gesuch erstellen
- Festpreis oder variabler Preis (abhängig von Personenzahl/Gewicht)
- Details: Von/Bis/Zwischenziele, Zeitraum, Fahrzeug/Anhänger, verfügbare Kapazitäten, Einschränkungen (z.B. keine Tiere, Nichtraucher), persönliche Hinweise
- Kommunikation und Zahlungsabwicklung integriert
- Speicherung aller Daten für Statistik

### Tracking
- Fahrer kann Standort teilen
- Statusabfrage der Fahrt möglich

### Chat
- Kommunikation zwischen Fahrer und Mitfahrer möglich
- Integration mit Fahrtdaten zur zeitlichen Zuordnung

---

## Nutzung

### Backend starten

```bash
# Repository klonen
git clone https://github.com/Konzepte-moderner-Softwareentwicklung/Backend.git
cd Backend
mkdir logging_data
chmod 777 logging_data

# Docker-Images bauen
docker compose build

# add you ip and port to frontend in nginx/nginx.conf
# start frontend

# Entwicklungsumgebung starten
docker compose up
