CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "name" varchar NOT NULL,
  "phone" varchar NOT NULL DEFAULT '',
  "password_hash" varchar NOT NULL DEFAULT '',
  "google_id" varchar NOT NULL DEFAULT '',
  "avatar_url" varchar NOT NULL DEFAULT '',
  "role" varchar NOT NULL DEFAULT 'user',
  "is_verified" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "destinations" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "country" varchar NOT NULL,
  "city" varchar NOT NULL,
  "description" text NOT NULL DEFAULT '',
  "image_url" varchar NOT NULL DEFAULT '',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "tour_packages" (
  "id" bigserial PRIMARY KEY,
  "title" varchar NOT NULL,
  "description" text NOT NULL DEFAULT '',
  "destination_id" bigint NOT NULL,
  "price" decimal(15,2) NOT NULL,
  "duration_days" int NOT NULL,
  "max_participants" int NOT NULL DEFAULT 20,
  "min_participants" int NOT NULL DEFAULT 1,
  "category" varchar NOT NULL DEFAULT 'adventure',
  "image_url" varchar NOT NULL DEFAULT '',
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "tour_itineraries" (
  "id" bigserial PRIMARY KEY,
  "tour_package_id" bigint NOT NULL,
  "day_number" int NOT NULL,
  "title" varchar NOT NULL,
  "description" text NOT NULL DEFAULT '',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "tour_images" (
  "id" bigserial PRIMARY KEY,
  "tour_package_id" bigint NOT NULL,
  "image_url" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "tour_facilities" (
  "id" bigserial PRIMARY KEY,
  "tour_package_id" bigint NOT NULL,
  "name" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "tour_schedules" (
  "id" bigserial PRIMARY KEY,
  "tour_package_id" bigint NOT NULL,
  "start_date" date NOT NULL,
  "end_date" date NOT NULL,
  "available_slots" int NOT NULL,
  "status" varchar NOT NULL DEFAULT 'available',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "bookings" (
  "id" bigserial PRIMARY KEY,
  "booking_code" varchar UNIQUE NOT NULL,
  "user_id" bigint NOT NULL,
  "tour_package_id" bigint NOT NULL,
  "tour_schedule_id" bigint,
  "travel_date" date NOT NULL,
  "num_participants" int NOT NULL,
  "total_price" decimal(15,2) NOT NULL,
  "status" varchar NOT NULL DEFAULT 'pending',
  "notes" text NOT NULL DEFAULT '',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "booking_participants" (
  "id" bigserial PRIMARY KEY,
  "booking_id" bigint NOT NULL,
  "name" varchar NOT NULL,
  "id_card_number" varchar NOT NULL DEFAULT '',
  "date_of_birth" date,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "payments" (
  "id" bigserial PRIMARY KEY,
  "booking_id" bigint UNIQUE NOT NULL,
  "payment_method" varchar NOT NULL DEFAULT 'bank_transfer',
  "amount" decimal(15,2) NOT NULL,
  "status" varchar NOT NULL DEFAULT 'pending',
  "payment_token" varchar NOT NULL DEFAULT '',
  "payment_url" varchar NOT NULL DEFAULT '',
  "paid_at" timestamptz,
  "expires_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "reviews" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "tour_package_id" bigint NOT NULL,
  "booking_id" bigint NOT NULL,
  "rating" int NOT NULL CHECK (rating >= 1 AND rating <= 5),
  "comment" text NOT NULL DEFAULT '',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL DEFAULT '',
  "client_ip" varchar NOT NULL DEFAULT '',
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- Foreign key constraints
ALTER TABLE "tour_packages" ADD FOREIGN KEY ("destination_id") REFERENCES "destinations" ("id");
ALTER TABLE "tour_itineraries" ADD FOREIGN KEY ("tour_package_id") REFERENCES "tour_packages" ("id") ON DELETE CASCADE;
ALTER TABLE "tour_images" ADD FOREIGN KEY ("tour_package_id") REFERENCES "tour_packages" ("id") ON DELETE CASCADE;
ALTER TABLE "tour_facilities" ADD FOREIGN KEY ("tour_package_id") REFERENCES "tour_packages" ("id") ON DELETE CASCADE;
ALTER TABLE "tour_schedules" ADD FOREIGN KEY ("tour_package_id") REFERENCES "tour_packages" ("id") ON DELETE CASCADE;
ALTER TABLE "bookings" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "bookings" ADD FOREIGN KEY ("tour_package_id") REFERENCES "tour_packages" ("id");
ALTER TABLE "bookings" ADD FOREIGN KEY ("tour_schedule_id") REFERENCES "tour_schedules" ("id");
ALTER TABLE "booking_participants" ADD FOREIGN KEY ("booking_id") REFERENCES "bookings" ("id") ON DELETE CASCADE;
ALTER TABLE "payments" ADD FOREIGN KEY ("booking_id") REFERENCES "bookings" ("id");
ALTER TABLE "reviews" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "reviews" ADD FOREIGN KEY ("tour_package_id") REFERENCES "tour_packages" ("id");
ALTER TABLE "reviews" ADD FOREIGN KEY ("booking_id") REFERENCES "bookings" ("id");
ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

-- Indexes
CREATE INDEX ON "tour_packages" ("destination_id");
CREATE INDEX ON "tour_packages" ("category");
CREATE INDEX ON "tour_packages" ("is_active");
CREATE INDEX ON "tour_itineraries" ("tour_package_id");
CREATE INDEX ON "tour_schedules" ("tour_package_id");
CREATE INDEX ON "tour_schedules" ("start_date");
CREATE INDEX ON "bookings" ("user_id");
CREATE INDEX ON "bookings" ("tour_package_id");
CREATE INDEX ON "bookings" ("status");
CREATE INDEX ON "booking_participants" ("booking_id");
CREATE INDEX ON "payments" ("booking_id");
CREATE INDEX ON "payments" ("status");
CREATE INDEX ON "reviews" ("tour_package_id");
CREATE INDEX ON "reviews" ("user_id");
CREATE INDEX ON "sessions" ("user_id");
