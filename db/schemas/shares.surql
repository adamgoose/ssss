DEFINE TABLE shares SCHEMAFULL;

DEFINE FIELD secret ON shares TYPE record<secrets>;
DEFINE FIELD user ON shares TYPE record<users>;
DEFINE FIELD key ON shares TYPE int;
DEFINE FIELD share ON shares TYPE string;
