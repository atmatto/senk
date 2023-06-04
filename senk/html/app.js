const add = (parent, tag, text = "", props = {}) => {
    const element = document.createElement(tag)
    element.textContent = text
    for (const prop of Object.entries(props)) {
        element[prop[0]] = prop[1]
    }
    parent.appendChild(element)
    return element
}

const buildIndex = (data) => {
    const main = document.getElementsByTagName("main")[0]
    const list = add(main, "ul")
    for (const note of data) {
        add(add(list, "li"), "a", "~" + note["Path"], {href: "/~" + note["Path"]})
    }
}

const getIndex = (user) => {
    const main = document.getElementsByTagName("main")[0]
    if (user === "") { // Get the index for the current user
        add(main, "p", "Viewing index")
        fetch("/api/index")
            .then(resp => {
                if (!resp.ok) {
                    // TODO: error handling
                    throw console.error("Couldn't get index", resp)
                }
                return resp.json()
            })
            .then(data => {
                buildIndex(data)
            })
    } else { // Get the index for the specified user
        // TODO: backend
        add(main, "p", "Viewing index of ~" + user)
    }
}

const buildEditor = (path, data) => {
    const main = document.getElementsByTagName("main")[0]
    add(main, "p", path.slice(1))
    /* const editor = */ add(main, "textarea", data)
}

const getNote = (user, id) => {
    const path = "/~" + user + "/" + id
    fetch(path + "/raw")
        .then(resp => {
            if (!resp.ok) {
                // TODO: error handling
                throw console.error("Couldn't get note", resp)
            }
            return resp.text()
        })
        .then(data => {
            buildEditor(path, data)
        })
}

window.onload = () => {
    const path = document.location.pathname.slice(1).split("/", 2)

    let elements = 0
    for (const e of path) {
        if (e !== "") {
            elements++
        }
    }
    
    switch (elements) {
    case 0:
    case 1:
        getIndex(path[0].slice(1))
        break
    case 2:
        getNote(path[0].slice(1), path[1])
        break
    }
}
