# Multi-site operation

Multi-site mode runs independent energy balance domains in one evcc process.
It is opt-in, and the existing single-site configuration and API remain
unchanged.

## Configuration

Use `sites` instead of `site` and give every loadpoint a stable name and exactly
one site assignment:

```yaml
sites:
  - name: home
    title: Home
    circuit: home-root
    tariffs:
      grid: home-grid-tariff
    meters:
      grid: home-grid
      pv: [home-pv]

  - name: office
    title: Office
    circuit: office-root
    tariffs:
      grid: office-grid-tariff
    meters:
      grid: office-grid
      pv: [office-pv]

  - name: cabin
    title: Cabin
    meters:
      grid: cabin-grid

loadpoints:
  - name: garage
    site: home
    charger: home-charger

  - name: parking
    site: office
    charger: office-charger

  - name: carport
    site: cabin
    charger: cabin-charger
```

An assignment may alternatively be declared as `loadpoints: [garage]` on a
site. Mixing `site` and `sites`, duplicate names, unknown references, ambiguous
assignments, and unassigned loadpoints are rejected at startup.

Vehicles are process-wide. Every configured vehicle is available for selection
and identification at every site. A shared coordinator ensures that one vehicle
can only belong to one loadpoint at a time, including across site boundaries.

Every site gets the same stable API prefix:

```text
/api/sites/home/...
/api/sites/office/...
```

WebSocket and `/api/state` updates from all sites are scoped below
`sites.<name>`. For compatibility, the first site's state and API are mirrored
at the existing root endpoints. Each site owns its meter set, loadpoints, control loop, power
balance, tariffs, root circuit, HEMS instance, and charging decisions. Vehicles
and their ownership are shared across sites.
MQTT uses a `sites/<name>` subtopic for every site and mirrors the first site at
the existing root topic. InfluxDB measurements
carry a `site` tag, and notification state is filtered to the originating site.

Charging sessions, energy history and HEMS grid sessions carry an explicit site
identity and are filtered by exact site name. Existing prefixed session and
energy-history records are migrated in place when their site starts.
The web UI offers a site selector and a read-only aggregate overview. Loadpoints
created in the configuration UI store their site assignment and join that site
after the normal configuration restart.

Define all devices once, add `sites`, and assign every loadpoint exactly once.
Startup rejects missing, duplicate, ambiguous, or cross-site assignments before
any control loop starts. The configured site names appear in `/api/state` as
`siteNames`.

All sites currently use the process-wide operating voltage, so they must share
that voltage setting. SHM, network listeners and authentication remain
process-wide integrations. Site-specific HEMS configuration belongs inside its
site entry; database-configured loadpoints can select a site in the UI.
