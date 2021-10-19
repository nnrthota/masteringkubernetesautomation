const {
    KubeConfig,
    Client
} = require('kubernetes-client')
const kubeconfig = new KubeConfig()


module.exports.getConfigFile = () => {
    return `${require('os').homedir()}/.kube/config`
}

module.exports.Client = (config) => {

    kubeconfig.loadFromFile(config)
    const Request = require('kubernetes-client/backends/request')

    const backend = new Request({
        kubeconfig
    })
    const client = new Client({
        backend,
        version: '1.13'
    })
    return client
}
