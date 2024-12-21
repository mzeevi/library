# Library

This repository contains a Go implementation of the following [exercise](https://github.com/dana-team/onboarding/tree/main/09-python).

## Run

### Set Up MongoDB

To run the application, you'll first need a `MongoDB` instance. You can easily set up a local `MongoDB` instance using the `Makefile`:

#### Start MongoDB

```bash
$ make setup-local-mongo
```

This command starts a MongoDB instance on `localhost:27017`.

#### Stop MongoDB

```bash
$ make teardown-local-mongo
```

This command stops and removes the MongoDB instance.

### Run the Application

After setting up `MongoDB`, use the `Makefile` to start the application. Example:

```bash
$ make run/api \
    DB_DSN=mongodb://localhost:27017 \
    JWT_SECRET=pei3einoh0Beem6uM6Ungohn2heiv5lah1ael4joopie5JaigeikoozaoTew2Eh6 \
    ADMIN_USER=admin \
    ADMIN_PASSWORD=admin \
    CREATE_ADMIN=true \
    DEMO_PATRONS=true \
    DEMO_BOOKS=true
```

In this example:
- `DB_DSN`: Specifies the MongoDB connection string.
- `JWT_SECRET`: A secret string used for signing JWTs.
- `ADMIN_USER` and `ADMIN_PASSWORD`: Credentials for the admin user.
- `CREATE_ADMIN`: Whether to create the admin user (`true` or `false`).
- `DEMO_PATRONS` and `DEMO_BOOKS`: Flags for wehther to create demo data.

## Build

To build the application as a Docker image, use the Makefile. Example:

```bash
$ make docker-build IMG=<name>:<tag>
```

Replace `<name>:<tag>` with the desired image name and tag.

## Test

This project uses [`testcontainers`](https://testcontainers.com/) for integration tests. A `MongoDB` container is spun up automatically during testing.

To run all tests:

```bash
$ make audit
```