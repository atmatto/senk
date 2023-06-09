const onLinkClick = (e) => {
    if (e.button === 0) {
        e.preventDefault();
        document.getElementById("status").classList.add("inactive")
        const path = e.target.pathname
        build(path)
        history.pushState(null, "", path)
    }
}

window.onpopstate = (e) => {
    document.getElementById("status").classList.add("inactive")
    build(document.location.pathname)
    e.preventDefault()
}

const add = (parent, tag, text = "", props = {}) => {
    const element = document.createElement(tag)
    element.textContent = text
    if (tag === "a" && props.href?.at?.(0) === "/" && !("onclick" in props)) {
        element.onclick = onLinkClick
    }
    for (const prop of Object.entries(props)) {
        element[prop[0]] = prop[1]
    }
    if (parent !== null) {
        parent.appendChild(element)
    }
    return element
}

const showError = (text) => {
    document.getElementById("statustext").textContent = text
    document.getElementById("status").classList.remove("inactive")
}

const buildIndex = (data) => {
    const main = document.getElementsByTagName("main")[0]
    main.replaceChildren([])
    const list = add(main, "ul")
    for (const note of data) {
        const path = "/~" + note["Path"]
        add(add(list, "li"), "a", "~" + note["Path"], {href: path})
    }
}

const getIndex = (user) => {
    const main = document.getElementsByTagName("main")[0]
    main.replaceChildren([])
    if (user === "") { // Get the index for the current user
        fetch("/api/index")
            .then(resp => {
                if (!resp.ok) {
                    // TODO: error handling
                    throw new Error(resp.status + " " + resp.statusText)
                }
                return resp.json()
            })
            .then(data => {
                buildIndex(data)
            })
            .catch(err => showError("Error getting index: " + err.message))
    } else { // Get the index for the specified user
        // TODO: backend
        showError("Feature not implemented.")
        add(main, "p", "Viewing index of " + user)
    }
}

const buildEditor = (path, data) => {
    const main = document.getElementsByTagName("main")[0]
    /* const editor = */ add(main, "textarea", data)
}

const getNote = (user, id) => {
    const main = document.getElementsByTagName("main")[0]
    main.replaceChildren([])
    const path = "/" + user + "/" + id
    fetch(path + "/raw")
        .then(resp => {
            if (!resp.ok) {
                // TODO: error handling
                throw new Error(resp.status + " " + resp.statusText)
            }
            return resp.text()
        })
        .then(data => {
            buildEditor(path, data)
        })
        .catch(err => showError("Error getting note: " + err.message))
}

const build = (path) => {
    path = path.slice(1).split("/", 2)
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

window.onload = () => {
    document.getElementById("senk").onclick = onLinkClick
    build(document.location.pathname)
}
