CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT,
  phone TEXT,
  password TEXT,
  created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE category (
  id SERIAL PRIMARY KEY,
  name TEXT
);

CREATE TABLE pharmacies (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  address TEXT,
  latitude TEXT,
  longitude TEXT,
  city TEXT,
  phone TEXT,
  created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE medications (
  id SERIAL PRIMARY KEY,
  status TEXT,
  commercial_status TEXT,
  speciality TEXT,
  dosage TEXT,
  form TEXT,
  presentation TEXT,
  pp TEXT,
  active_substance TEXT,
  therapeutic_class TEXT,
  epi TEXT,
  ppv TEXT,
  ph TEXT,
  pfht TEXT,
  code TEXT,
  tva TEXT, 
  created_at TIMESTAMP DEFAULT now(),
  description TEXT,
  common_sd TEXT[],
  serious_sd TEXT[], 
  general_info TEXT[],
  category_id INTEGER,
  FOREIGN KEY (category_id) REFERENCES category(id) ON DELETE CASCADE
);

CREATE TABLE stock (
  id SERIAL PRIMARY KEY,
  pharmacy_id INT REFERENCES pharmacies(id) ON DELETE CASCADE,
  medication_id INT REFERENCES medications(id) ON DELETE CASCADE,
  quantity INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP DEFAULT now(),
  UNIQUE (pharmacy_id, medication_id)
);


