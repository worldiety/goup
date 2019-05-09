//
//  ViewController.swift
//  iosApp
//
//  Created by Christian Maier on 08.05.19.
//  Copyright © 2019 Worldiety. All rights reserved.
//

import Myproject
import UIKit

/// Helper class to wrap a closure as MyprojectPkgaHelloCallbackProtocol
class Callback: NSObject, MyprojectPkgaHelloCallbackProtocol {
    private let callback: () -> String
    init(_ callback: @escaping () -> String) {
        self.callback = callback
        super.init()
    }

    func yourName() -> String {
        return callback()
    }
}

class ViewController: UIViewController {
    override func viewDidLoad() {
        super.viewDidLoad()

        /// Call a exported go function
        let goString = MyprojectMyprojectAnExportedProjectLevelFunc()
        print(goString)

        guard let map2 = MyprojectPkgbGetMap2(), let keys = map2.keys() else {
            fatalError("Something's Not Quite Right...")
        }

        for index in 0..<keys.len() {
            let name = keys.get(index, error: nil)
            let detail = map2.get(name)
            print("\(index) - \(name): \(detail)")
        }

        let lastName = keys.len() > 0 ? keys.get(keys.len() - 1, error: nil) : ""

        /// Call a exported go function with an callback paramer
        let goCallback = MyprojectPkgaNiceCallback(Callback {
            lastName
        })
        print(goCallback)
    }
}
