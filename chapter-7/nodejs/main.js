config = require("./helper")
dep = require("./deployments")
svc = require("./service")
pvc = require("./pvc")
secret = require("./secret")
ns = require("./namespace")
msqldepManifest = require("./static/msql-dep.json")
msqlsvcManifest = require("./static/mysql-svc.json")
msqlpvcManifest = require("./static/mysql-pvc.json")
secretmanifest = require("./static/secret.json")
wordpresspvcManifest = require("./static/wordpress-pvc.json")
wordpressdepManifest = require("./static/wordpress-dep.json")
wordpresssvcManifest = require("./static/wordpress-svc.json")

async function main(client) {
    try {
        client = config.Client(await config.getConfigFile())
        /* CREATE */
        dep.create(client, "default", msqldepManifest)
        dep.create(client, "default", wordpressdepManifest)
        svc.create(client, "default", msqlsvcManifest)
        svc.create(client, "default", wordpresssvcManifest)
        pvc.create(client, "default", msqlpvcManifest)
        pvc.create(client, "default", wordpresspvcManifest)
        secret.create(client, "default", secretmanifest)

        /* LIST */
        dep.list(client,"default")
        secret.list(client,"default")
        pvc.list(client,"default")
        svc.list(client,"default")
        ns.list(client)

    } catch (err) {
        console.error('Error: ', err)
    }

}

main()