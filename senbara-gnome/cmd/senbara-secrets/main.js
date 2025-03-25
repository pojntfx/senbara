import GLib from "gi://GLib";
import Secret from "gi://Secret";

const schema = new Secret.Schema(
  "com.pojtinger.felicitas.Senbara",
  Secret.SchemaFlags.NONE,
  { type: Secret.SchemaAttributeType.STRING }
);

const loop = GLib.MainLoop.new(null, false);

Secret.password_store(
  schema,
  { type: "id_token" },
  Secret.COLLECTION_DEFAULT,
  "id_token",
  "example-id-token-value",
  null,
  () =>
    Secret.password_lookup(schema, { type: "id_token" }, null, (_, res) => {
      print(Secret.password_lookup_finish(res));

      loop.quit();
    })
);

loop.run();
