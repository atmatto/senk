const goto = (path) => {
    document.getElementById("status").classList.add("inactive")
    build(path)
    history.pushState(null, "", path)
}

const onLinkClick = (e) => {
    if (e.button === 0) {
        e.preventDefault();
        goto(e.target.pathname)
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

const buildIndex = (data, trash = false) => {
    const main = document.getElementsByTagName("main")[0]
    main.replaceChildren([])
    const list = add(main, "ul")
    for (const note of data) {
        const path = (trash ? "/trash" : "") + "/~" + note["Path"]
        add(add(list, "li"), "a", "~" + note["Path"], {href: path})
    }
    if (!trash)
        add(add(list, "li"), "a", "Trash", {href: "/trash"})
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
        fetch("/api/index/" + user)
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
    }
}

const getTrash = () => {
    const main = document.getElementsByTagName("main")[0]
    main.replaceChildren([])
    fetch("/api/trash")
        .then(resp => {
            if (!resp.ok) {
                // TODO: error handling
                throw new Error(resp.status + " " + resp.statusText)
            }
            return resp.json()
        })
        .then(data => {
            buildIndex(data, true)
        })
        .catch(err => showError("Error getting index: " + err.message))
}

const buildNoteDetails = (parent, path) => {
    const details = add(parent, "details", "", {id: "details"})
    add(details, "summary", "Manage note")
    const list = add(details, "ul")
    add(add(list, "li"), "button", "Delete", {onclick: () => {
        fetch(path, {method: "DELETE"})
            .then(resp => {
                if (!resp.ok) {
                    throw new Error(resp.status + " " + resp.statusText)
                }
                goto("/trash")
                showError("Note deleted")
            })
            .catch(err => showError("Error deleting note: " + err.message))
    }})
}

const buildEditor = (path, data) => {
    const main = document.getElementsByTagName("main")[0]
    buildNoteDetails(main, path)
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
    path = path.slice(1).split("/") // slice strips leading slash
    let elements = 0
    for (const e of path) {
        if (e !== "") {
            elements++
        }
    }

    const header = document.getElementsByTagName("header")[0]
    const title = document.getElementById("title")
    if (path[0] === "trash") {
        console.log(elements, path)
        switch (elements - 1) {
        case 0:
            getTrash()
            title.replaceChildren(add(null, "span", "Trash"))
            header.classList.remove("notitle")
            break
        case 2:
            // TODO: Deleted note viewer
            showError("Error: Not implemented")
            header.classList.add("notitle")
            break
        }
    } else {
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
}

window.onload = () => {
    document.getElementById("senk").onclick = onLinkClick
    build(document.location.pathname)
}
