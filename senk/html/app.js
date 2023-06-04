const add = (parent, tag, text = "", props = {}) => {
    const element = document.createElement(tag)
    element.textContent = text
    for (const prop of Object.entries(props)) {
        element[prop[0]] = prop[1]
    }
    if (parent !== null) {
        parent.appendChild(element)
    }
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
        add(main, "p", "Viewing index of " + user)
    }
}

const buildEditor = (path, data) => {
    const main = document.getElementsByTagName("main")[0]
    /* const editor = */ add(main, "textarea", data)
}

const getNote = (user, id) => {
    const path = "/" + user + "/" + id
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

    const header = document.getElementsByTagName("header")[0]
    const title = document.getElementById("title")
    switch (elements) {
    case 0:
        getIndex("")
        header.classList.add("notitle")
        break
    case 1:
        getIndex(path[0])
        title.replaceChildren(add(null, "span", path[0]))
        header.classList.remove("notitle")
        break
    case 2:
        getNote(path[0], path[1])
        title.replaceChildren(add(null, "a", path[0], {href: "/"+path[0]}), add(null, "span", "/"+path[1]))
        header.classList.remove("notitle")
        break
    }
}
